package client

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"testing"
)

func setupLogsCapture() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return zap.New(core), logs
}

func TestClient(t *testing.T) {
	logger, logs := setupLogsCapture()

	clientConfig := &ClientConfig{"test-client",
		"localhost:8080",
		500,
		0,
		0,
		10000,
		nil}

	client := &Client{clientConfig, logger}

	ctx := context.Background()
	err := client.Listen(ctx)

	assert.True(t, err == nil)
	assert.True(t, logs.Len() > 0)
}
