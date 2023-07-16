package logging

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewDefaultLogger() *zap.Logger {
	return zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "@timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			EncodeDuration: zapcore.NanosDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zapcore.InfoLevel),
	), zap.AddCaller(), zap.AddCallerSkip(1))
}

func NewCombinedLogger(filename string) *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	logFile, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger
}

func FailIfErr(err error, msg string) {
	if err != nil {
		log.Fatal(msg, zap.Error(err))
	}
}

/*
// WithFields returns a new context with the provided fields attached as metadata to future loggers.
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	if len(fields) == 0 {
		return ctx
	}
	return With(ctx, From(ctx).With(fields...))
}

// Sync flushes any buffered log entries.
func Sync(ctx context.Context) error {
	return From(ctx).Sync()
}
*/
