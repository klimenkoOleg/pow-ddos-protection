package server

import (
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net"
	"strings"
)

type BookQuote struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
}

type Book struct {
	quotes []*BookQuote
}

func NewBook(jsonQuotes []byte) (*Book, error) {
	var quotes []*BookQuote
	if err := json.Unmarshal(jsonQuotes, &quotes); err != nil {
		return nil, err
	}
	return &Book{quotes: quotes}, nil
}

func (b *Book) GetRandQuote() *BookQuote {
	i := rand.Intn(len(b.quotes))
	return b.quotes[i]
}

func (b *Book) ServeRequest(conn net.Conn, requestLog zap.Logger) {
	requestLog.Info("write response")

	q := b.GetRandQuote()
	r := strings.NewReader(q.Quote)
	_, err := io.Copy(conn, r)
	if err != nil {
		requestLog.Warn("failed to write response")
	}
}
