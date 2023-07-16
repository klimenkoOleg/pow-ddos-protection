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
)

type Client struct {
	Cfg *ClientConfig
	Log zap.Logger
}

// old startFetchWorkers
func (c *Client) Listen(ctx context.Context) error {
	// TODO mpve to client
	creationPause := 500 //conf.Timeout / time.Duration(conf.FetchWorkers)

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ { //conf.FetchWorkers; i++ {
		go func() {
			wg.Add(1)
			c.runFetchWorker(ctx, "localhost:8080", 500, c.Log, i)
		}()
		time.Sleep(time.Duration(creationPause))
	}

	wg.Wait()

	return nil
}

func (c *Client) runFetchWorker(ctx context.Context, serverAddr string, timeout int, log zap.Logger, id int) {
	log.Info("worked started", zap.Int("id", id))

	// TODO get number of iterations from config
	for i := 0; i < 10; i++ {
		select {
		// TODO is it needed?
		case <-ctx.Done():
			return
		default:
			fetchQuote(log, serverAddr, c.Cfg.PrivateKey, c.Cfg.HashcashMaxIterations)
			time.Sleep(time.Duration(timeout))

		}
	}
}

func fetchQuote(log zap.Logger, serverAddr string, privateKey *rsa.PrivateKey, hashcashMaxIterations int) string {
	log.Info("connecting to server", zap.String("server_addr", serverAddr))
	// TODO probably open conneection a leve higher
	conn, err := net.Dial("tcp", serverAddr)
	defer conn.Close()
	CheckErr(log, err)

	// 1. requesting challenge
	msg := &message.Message{Header: message.Step1ChallengeRequest}
	encryptedMsg, err := message.EncodeRSAGob(msg, privateKey.PublicKey)
	CheckErr(log, err)
	_, err = conn.Write(encryptedMsg)
	CheckErr(log, err)

	// reading and parsing response
	received := make([]byte, 1024)
	n, err := conn.Read(received)
	CheckErr(log, err)
	receivedDecreptedMsg, err := message.DecodeRSAGob(received[:n], *privateKey)
	CheckErr(log, err)

	var hashcash pow.HashcashData
	err = json.Unmarshal(receivedDecreptedMsg.Payload, &hashcash)
	CheckErr(log, err)
	log.Info("PoW task from server", zap.String("hashcash", hashcash.Stringify()))

	// 2. got challenge, compute hashcash
	// TODO
	hashcash, err = hashcash.ComputeHashcash(hashcashMaxIterations)
	CheckErr(log, err)
	log.Info("PoW computed", zap.String("hashcash", hashcash.Stringify()))

	// marshal solution to json
	hashcashBytes, err := json.Marshal(hashcash)
	CheckErr(log, err)

	// 3. send challenge solution back to server
	solutionMsg := &message.Message{Header: message.Step3QuoteRequest, Payload: hashcashBytes}
	encryptedMsg, err = message.EncodeRSAGob(solutionMsg, privateKey.PublicKey)
	CheckErr(log, err)
	_, err = conn.Write(encryptedMsg)
	CheckErr(log, err)
	log.Info("PoW solution sent to server")

	// 4. get result quote from server
	n, err = conn.Read(received)
	CheckErr(log, err)
	receivedDecreptedMsg, err = message.DecodeRSAGob(received[:n], *privateKey)
	CheckErr(log, err)

	citation := string(receivedDecreptedMsg.Payload)
	log.Info("wisdom citation", zap.String("citation", citation))

	return citation
}

func CheckErr(log zap.Logger, err error) {
	if err != nil {
		log.Fatal("client error", zap.Error(err))
	}
}
