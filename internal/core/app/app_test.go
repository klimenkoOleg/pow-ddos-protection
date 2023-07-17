package app

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"pow-ddos-protection/internal/client"
	"pow-ddos-protection/internal/core/logging"
	"testing"
)

func setupLogsCapture() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return zap.New(core), logs
}

func TestApp(t *testing.T) {

	logger, logs := setupLogsCapture()
	ctx := context.Background()

	ctx = context.WithValue(ctx, "logger", logger)
	clientApp := NewApp(ctx, logging.NewDefaultLogger(), nil)

	clientApp.Start(startTester)

	assert.True(t, logs.Len() > 0)
}

// startClient - to be invoked on the App start, run in a goroutine as a Listener
func startTester(ctx context.Context, app *App) ([]Listener, error) {
	log := ctx.Value("logger").(*zap.Logger)
	log.Info("test")
	clientConfig := &client.ClientConfig{"test-client",
		"localhost:8080",
		500,
		0,
		0,
		10000,
		nil}
	c := &client.Client{clientConfig, log}

	// Start sending requests for the book quotes
	return []Listener{
		c,
	}, nil
}

func CheckError(e error, t *testing.T) {
	if e != nil {
		t.Fail()
	}
}
