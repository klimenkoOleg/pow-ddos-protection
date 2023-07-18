package server

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"testing"
)

func setupLogsCapture() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return zap.New(core), logs
}

func TestApp(t *testing.T) {
	logger, _ := setupLogsCapture()

	cfg := &ServerConfig{"test-server", "8080", 5, 500, nil}
	server, err := New(cfg, logger)

	assert.True(t, err == nil)
	assert.True(t, server != nil)

}
