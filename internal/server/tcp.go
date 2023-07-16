package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"math/rand"
	"net"
	"pow-ddos-protection/internal/core/config"
	"pow-ddos-protection/internal/core/errors"
	"pow-ddos-protection/internal/message"
	"pow-ddos-protection/internal/pow"
	"strconv"
	"time"
)

const (
	// ErrAddRoutes is the error returned when adding routes to the router fails.
	ErrAddRoutes = errors.Error("failed to add routes")
	// ErrServer is the error returned when the server stops due to an error.
	ErrServer = errors.Error("listen stopped with error")
)

// Quotes - const array of quotes to respond on client's request
var Quotes = []string{
	"All saints who remember to keep and do these sayings, " +
		"walking in obedience to the commandments, " +
		"shall receive health in their navel and marrow to their bones",

	"And shall find wisdom and great treasures of knowledge, even hidden treasures",

	"And shall run and not be weary, and shall walk and not faint",

	"And I, the Lord, give unto them a promise, " +
		"that the destroying angel shall pass by them, " +
		"as the children of Israel, and not slay them",
}

// Config represents the configuration of the http listener.
//type Config struct {
//	TcpAddress string `yaml:"tcp-address"`
//}

type Server struct {
	//conf *Config
	cfg  *config.Config
	log  *zap.Logger
	sock net.Listener
	//handler func(net.Conn, zap.Logger)
	//powReceive pow.Receiver
}

func New(cfg *config.Config, log *zap.Logger) (*Server, error) {
	port := cfg.AppConfig.Port
	log.Info(fmt.Sprintf("tcp server starting on address: %s", port))

	socket, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	s := &Server{
		cfg:  cfg,
		log:  log,
		sock: socket,
		//handler:    handler,
		//powReceive: pow.NewReceiver(conf.Difficulty, conf.ProofTokenSize),
	}
	//s := &Server{
	//	conf:       conf,
	//	log:        log,
	//	sock:       socket,
	//	handler:    handler,
	//powReceive: pow.NewReceiver(conf.Difficulty, conf.ProofTokenSize),
	//}
	//go s.listen()
	//return s, nil
	return s, nil
}

func (s *Server) Close() error {
	return s.sock.Close()
}

// his is needed for APP design!!!
func (s *Server) Listen(ctx context.Context) error {
	/*s.log.Info(fmt.Sprintf("http server starting on port: %s", s.port))

	err := s.server.ListenAndServe()
	if err != nil {
		return ErrServer.Wrap(err)
	}

	logging.From(ctx).Info("http server stopped")

	return nil*/
	for {
		conn, err := s.sock.Accept()
		connID := uuid.New()
		s.log.Info("Accepted new connection", zap.String("connID", connID.String()))
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			s.log.Warn("failed to listen socket", zap.String("connID", connID.String()))
			continue
		}
		go s.serveConn(conn, connID)
	}

}

func (s *Server) serveConn(conn net.Conn, connID uuid.UUID) {
	defer conn.Close()

	s.log.Info("Processing an incoming connection",
		zap.String("connID", connID.String()),
		zap.String("remote address", conn.RemoteAddr().String()))

	// 1. Process Request for challenge
	received := make([]byte, 1024)
	n, err := conn.Read(received)
	CheckErr(s.log, err)

	receivedDecreptedMsg, err := message.DecodeRSAGob(received[:n], *s.cfg.PrivateKey)
	CheckErr(s.log, err)

	s.log.Info("Step 1 message header: " + strconv.Itoa(receivedDecreptedMsg.Header))

	if receivedDecreptedMsg.Header != message.Step1ChallengeRequest {
		s.log.Error(fmt.Sprintf("Incorrect state from client, expected %v, got %v",
			message.Step1ChallengeRequest,
			receivedDecreptedMsg.Header))
	}

	randValue := rand.Intn(100000)
	hashcash := pow.HashcashData{
		Version:    1,
		ZerosCount: s.cfg.AppConfig.HashcashZerosCount,
		Date:       time.Now().Unix(),
		Resource:   conn.RemoteAddr().String(),
		Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", randValue))),
		Counter:    0,
	}

	hashcashEncoded, err := json.Marshal(hashcash)
	s.log.Info("hashcashEncoded: " + hashcash.Stringify())
	CheckErr(s.log, err)
	challengeResponseMessage := &message.Message{message.Step2TwoChallengeResponse, hashcashEncoded}
	privKey := *s.cfg.PrivateKey
	encryptedMsg, err := message.EncodeRSAGob(challengeResponseMessage, privKey.PublicKey)
	CheckErr(s.log, err)
	//s.log.Info("challengeResponseMessage: " + strconv.Itoa(challengeResponseMessage.Header) + " : " + challengeResponseMessage.Payload)

	conn.Write(encryptedMsg)

	// 2. Get hashcash solution from client
	n, err = conn.Read(received)
	CheckErr(s.log, err)

	receivedDecreptedMsg, err = message.DecodeRSAGob(received[:n], *s.cfg.PrivateKey)
	CheckErr(s.log, err)

	if receivedDecreptedMsg.Header != message.Step3QuoteRequest {
		s.log.Error(fmt.Sprintf("Incorrect state from client, expected %v, got %v",
			message.Step1ChallengeRequest,
			receivedDecreptedMsg.Header))
	}

	//var hashcash pow.HashcashData
	err = json.Unmarshal(receivedDecreptedMsg.Payload, &hashcash)
	CheckErr(s.log, err)

	// validate hashcash params
	if hashcash.Resource != conn.RemoteAddr().String() {
		s.log.Error("invalid hashcash resource")
		return
	}

	// decoding rand from base64 field in received client's hashcash
	randValueBytes, err := base64.StdEncoding.DecodeString(hashcash.Rand)
	CheckErr(s.log, err)
	randValueFromClient, err := strconv.Atoi(string(randValueBytes))
	CheckErr(s.log, err)

	if randValue != randValueFromClient {
		s.log.Error("invalid randValue from client")
		return
	}

	// sent solution should not be outdated
	if time.Now().Unix()-hashcash.Date > int64(s.cfg.AppConfig.HashcashTimeout) {
		s.log.Error("hashcash texpired")
		return
	}

	//to prevent indefinite computing on server if client sent hashcash with 0 counter
	maxIter := hashcash.Counter
	if maxIter == 0 {
		maxIter = 1
	}
	_, err = hashcash.ComputeHashcash(maxIter)
	CheckErr(s.log, err)

	//msg := protocol.Message{protocol
	//	Header:  protocol.ResponseResource,
	//	Payload: Quotes[rand.Intn(4)],
	//}

	quoteMsg := &message.Message{message.Step4QuoteResponse, []byte(Quotes[rand.Intn(4)])}
	//privKey := *s.cfg.PrivateKey
	quoteResponseMsg, err := message.EncodeRSAGob(quoteMsg, privKey.PublicKey)
	CheckErr(s.log, err)

	conn.Write(quoteResponseMsg)

	//

	/*// incoming request
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	receivedDecryptedMsg, err := message.DecodeRSAGob(buffer[:n], *s.cfg.PrivateKey)
	CheckErr(s.log, err)*/

	/*s.log.Info("Got message:" + strconv.Itoa(receivedDecryptedMsg.Header) + " : " + string(receivedDecryptedMsg.Payload))

	// write data to response
	time := time.Now().Format(time.ANSIC)
	responseStr := fmt.Sprintf("Your message is: %s. Received time: %v", receivedDecryptedMsg.Payload, time)
	s.log.Info("MESSAGE SENDING: " + responseStr)

	responseMsg := &message.Message{2, []byte(responseStr)}
	privKey := *s.cfg.PrivateKey
	encryptedMsg, err := message.EncodeRSAGob(responseMsg, privKey.PublicKey)
	CheckErr(s.log, err)

	conn.Write(encryptedMsg)*/

	//checkDuration, err := s.powReceive(conn)
	//if err != nil {
	//	log.Warn().Err(err).Dur("check_duration", checkDuration).Msg("refuse conn")
	//	return
	//}
	//log.Debug().
	//	Int("difficulty", int(s.conf.Difficulty)).
	//	Dur("check_duration", checkDuration).
	//	Msg("is valid proof")

	// TODO
	//s.handler(conn, log)
}

func CheckErr(log *zap.Logger, err error) {
	if err != nil {
		log.Fatal("client error", zap.Error(err))
	}
}

/*
	func StartServer(conf *Config, log zap.Logger, handler func(net.Conn, zap.Logger)) (*Server, error) {
		socket, err := net.Listen("tcp", conf.TCPAddress)
		if err != nil {
			return nil, err
		}
		s := &Server{
			conf:       conf,
			log:        log,
			sock:       socket,
			handler:    handler,
			//powReceive: pow.NewReceiver(conf.Difficulty, conf.ProofTokenSize),
		}
		go s.listen()
		return s, nil
	}
*/

/*
func (s *Server) listen() {
	for i := 0; ; i++ {
		conn, err := s.sock.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			s.log.Warn().Err(err).Msg("failed to listen socket")
			continue
		}
		go s.serveConn(conn, i)
	}
}

func (s *Server) serveConn(conn net.Conn, connID int) {
	defer conn.Close()

	log := s.log.With().
		Int("id", connID).
		Str("addr", conn.RemoteAddr().String()).
		Logger()
	log.Trace().Msg("receive conn")

	checkDuration, err := s.powReceive(conn)
	if err != nil {
		log.Warn().Err(err).Dur("check_duration", checkDuration).Msg("refuse conn")
		return
	}
	log.Debug().
		Int("difficulty", int(s.conf.Difficulty)).
		Dur("check_duration", checkDuration).
		Msg("is valid proof")

	s.handler(conn, log)
}
*/
