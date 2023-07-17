package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math/rand"
	"net"
	"pow-ddos-protection/internal/message"
	"pow-ddos-protection/internal/pow"
	"strconv"
	"time"
)

// WisdomResponses - const array of quotes to respond on client's request
var WisdomResponses = []string{
	"Love uprightness you who are rulers on earth",
	"Wisdom will never enter the soul of a wrong-doer, nor dwell in a body enslaved to sin",
	"Wisdom is a spirit friendly to humanity, though she will not let a blasphemer's words go unpunished",
	"For the spirit of the Lord fills the world, and that which holds everything together knows every word said",
	"No one who speaks what is wrong will go undetected, nor will avenging Justice pass by such a one",
	"There is a jealous ear that overhears everything, not even a murmur of complaint escapes it",
	"Do not court death by the errors of your ways, nor invite destruction through the work of your hands",
}

type Server struct {
	cfg  *ServerConfig
	log  *zap.Logger
	sock net.Listener
}

func New(cfg *ServerConfig, log *zap.Logger) (*Server, error) {
	port := cfg.Port
	log.Info(fmt.Sprintf("tcp server starting on address: %s", port))

	socket, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	s := &Server{
		cfg:  cfg,
		log:  log,
		sock: socket,
	}

	return s, nil
}

func (s *Server) Close() error {
	return s.sock.Close()
}

// Listen is called by App inside a goroutine
func (s *Server) Listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := s.sock.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return nil
				}
				s.log.Warn("failed to listen socket")
				continue
			}
			connID := uuid.New()
			s.log.Info("Accepted new connection", zap.String("connID", connID.String()))
			go s.serveConn(conn, connID)
		}
	}
}

func (s *Server) serveConn(conn net.Conn, connID uuid.UUID) {
	defer conn.Close()

	s.log.Info("Processing an incoming connection",
		zap.String("connID", connID.String()),
		zap.String("remote address", conn.RemoteAddr().String()))

	// 1. Get Request for challenge
	received := make([]byte, 1024)
	n, err := conn.Read(received)
	CheckErrOk(s.log, err)
	// decrypt RSA, decode by gob
	receivedDecreptedMsg, err := message.DecodeRSAGob(received[:n], *s.cfg.PrivateKey)
	CheckErrOk(s.log, err)
	s.log.Debug("Step 1 message header: " + strconv.Itoa(receivedDecreptedMsg.Header))

	if receivedDecreptedMsg.Header != message.Step1ChallengeRequest {
		s.log.Error(fmt.Sprintf("Incorrect state from client, expected %v, got %v",
			message.Step1ChallengeRequest,
			receivedDecreptedMsg.Header))
	}
	// prepare PoW challenge for clien
	randValue := rand.Intn(100000)
	hashcash := pow.HashcashData{
		Version:    1,
		ZerosCount: s.cfg.HashcashZerosCount,
		Date:       time.Now().Unix(),
		Resource:   conn.RemoteAddr().String(),
		Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", randValue))),
		Counter:    0,
	}
	// encrypt PoW challenge into Payload bytes
	hashcashEncoded, err := json.Marshal(hashcash)
	s.log.Debug("hashcashEncoded: " + hashcash.Stringify())
	CheckErrOk(s.log, err)
	challengeResponseMessage := &message.Message{message.Step2TwoChallengeResponse, hashcashEncoded}
	privKey := *s.cfg.PrivateKey
	// and encrypt it by RSA, encode it by gob for sending
	encryptedMsg, err := message.EncodeRSAGob(challengeResponseMessage, privKey.PublicKey)
	CheckErrOk(s.log, err)
	// sent to client via TCP
	conn.Write(encryptedMsg)
	s.log.Debug("Sent PoW task")

	// 2. Get hashcash solution from client
	n, err = conn.Read(received)
	CheckErrOk(s.log, err)
	s.log.Debug("Got solution")

	// decrypt RSA, decode by gob
	receivedDecreptedMsg, err = message.DecodeRSAGob(received[:n], *s.cfg.PrivateKey)
	CheckErrOk(s.log, err)
	s.log.Debug("Solution decoded : ", zap.Int("header", receivedDecreptedMsg.Header))

	if receivedDecreptedMsg.Header != message.Step3QuoteRequest {
		s.log.Error(fmt.Sprintf("Incorrect state from client, expected %v, got %v",
			message.Step1ChallengeRequest,
			receivedDecreptedMsg.Header))
	}
	// decode Payload into hashcash solution
	err = json.Unmarshal(receivedDecreptedMsg.Payload, &hashcash)
	CheckErrOk(s.log, err)
	s.log.Debug("Solution decoded : ", zap.Int("header", receivedDecreptedMsg.Header))

	// validate hashcash params
	if hashcash.Resource != conn.RemoteAddr().String() {
		s.log.Error("invalid hashcash resource")
		return
	}
	// decoding rand from base64 field in received client's hashcash
	randValueBytes, err := base64.StdEncoding.DecodeString(hashcash.Rand)
	CheckErrOk(s.log, err)
	randValueFromClient, err := strconv.Atoi(string(randValueBytes))
	CheckErrOk(s.log, err)

	if randValue != randValueFromClient {
		s.log.Error("invalid randValue from client")
		return
	}
	// sent solution should not be outdated
	if time.Now().Unix()-hashcash.Date > int64(s.cfg.HashcashTimeout) {
		s.log.Error("hashcash texpired")
		return
	}

	//to prevent indefinite computing on server if client sent hashcash with 0 counter
	maxIter := hashcash.Counter
	if maxIter == 0 {
		maxIter = 1
	}
	_, err = hashcash.ComputeHashcash(maxIter)
	CheckErrOk(s.log, err)

	// here all hashcode checks are passed, so select quote and send it to client
	quoteMsg := &message.Message{message.Step4QuoteResponse, []byte(WisdomResponses[rand.Intn(7)])}
	quoteResponseMsg, err := message.EncodeRSAGob(quoteMsg, privKey.PublicKey)
	CheckErrOk(s.log, err)
	s.log.Debug("Quote sending")
	// send the quote to client
	conn.Write(quoteResponseMsg)
}

func CheckErrOk(log *zap.Logger, err error) bool {
	if err != nil {
		log.Error("communication error", zap.Error(err))
		return false
	}
	return true
}
