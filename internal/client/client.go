package client

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"go.uber.org/zap"
	"net"
	"pow-ddos-protection/internal/message"
	"pow-ddos-protection/internal/pow"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	Cfg *ClientConfig
	Log *zap.Logger
}

// Listen invoked by App in a goroutine
func (c *Client) Listen(ctx context.Context) error {
	creationPause := c.Cfg.RequestsCreationTimeout

	wg := sync.WaitGroup{}

	for i := 0; i < c.Cfg.NumberOfClients; i++ {
		go func() {
			wg.Add(1)
			c.runClient(ctx, c.Cfg.ServerAddress, 500, c.Log, i)
		}()
		time.Sleep(time.Duration(creationPause))
	}

	wg.Wait()

	return nil
}

func (c *Client) runClient(ctx context.Context, serverAddr string, timeout int, log *zap.Logger, id int) {
	log.Info("worked started", zap.Int("id", id))

	for i := 0; i < c.Cfg.RequestsPerClient; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			quote, err := runQuoteWorkflow(log, serverAddr, c.Cfg.PrivateKey, c.Cfg.HashcashMaxIterations)
			if err != nil {
				log.Error("Client could not get quote", zap.Error(err))
			} else {
				log.Info("Got the quote, use it wisely", zap.String("quote", quote))
			}
			time.Sleep(time.Duration(timeout))
		}
	}
}

func runQuoteWorkflow(log *zap.Logger, serverAddr string, privateKey *rsa.PrivateKey, hashcashMaxIterations int) (string, error) {
	log.Info("connecting to server", zap.String("server_addr", serverAddr))
	conn, err := net.Dial("tcp", serverAddr)
	defer conn.Close()
	if err != nil {
		return "", errors.Wrap(err, "Failed connect to server address: "+serverAddr)
	}

	log.Info("connected to server", zap.String("server_addr", serverAddr))

	// 1. requesting challenge
	msg := &message.Message{Header: message.Step1ChallengeRequest}
	encryptedMsg, err := message.EncodeRSAGob(msg, privateKey.PublicKey)
	if err != nil {
		return "", errors.Wrap(err, "Failed requesting challenge")
	}
	_, err = conn.Write(encryptedMsg)
	if err != nil {
		return "", errors.Wrap(err, "Failed sent challenge to client")
	}

	// reading and parsing response
	received := make([]byte, 1024)
	n, err := conn.Read(received)
	if err != nil {
		return "", errors.Wrap(err, "Failed get challenge answer")
	}
	receivedDecreptedMsg, err := message.DecodeRSAGob(received[:n], *privateKey)
	if err != nil {
		return "", errors.Wrap(err, "Failed decode challenge answer")
	}

	// unwrap PoW task from server
	var hashcash pow.HashcashData
	err = json.Unmarshal(receivedDecreptedMsg.Payload, &hashcash)
	if err != nil {
		return "", errors.Wrap(err, "Failed extract Payload")
	}
	log.Info("PoW task from server", zap.String("hashcash", hashcash.Stringify()))

	// 2. PoW challenge: having the challenge, compute hashcash
	hashcash, err = hashcash.ComputeHashcash(hashcashMaxIterations)
	if err != nil {
		return "", errors.Wrap(err, "Failed validate Hashcash")
	}
	log.Info("PoW computed", zap.String("hashcash", hashcash.Stringify()))

	// marshal solution to json
	hashcashBytes, err := json.Marshal(hashcash)
	if err != nil {
		return "", errors.Wrap(err, "Failed marshal hashcash")
	}

	// 3. send challenge solution back to server
	solutionMsg := &message.Message{Header: message.Step3QuoteRequest, Payload: hashcashBytes}
	encryptedMsg, err = message.EncodeRSAGob(solutionMsg, privateKey.PublicKey)
	if err != nil {
		return "", errors.Wrap(err, "Failed encrypt PoW solution")
	}
	_, err = conn.Write(encryptedMsg)
	if err != nil {
		return "", errors.Wrap(err, "Failed sent PoW solution")
	}
	log.Info("PoW solution sent to server")

	// 4. get result quote from server
	n, err = conn.Read(received)
	if err != nil {
		return "", errors.Wrap(err, "Failed recaive PoW solution")
	}
	receivedDecreptedMsg, err = message.DecodeRSAGob(received[:n], *privateKey)
	if err != nil {
		return "", errors.Wrap(err, "Failed decode server quote")
	}

	citation := string(receivedDecreptedMsg.Payload)
	log.Info("wisdom citation", zap.String("citation", citation))

	return citation, nil
}
