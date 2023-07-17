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
		zap.NewAtomicLevelAt(zapcore.DebugLevel),
	), zap.AddCaller(), zap.AddCallerSkip(1))
}

func FailIfErr(err error, msg string) {
	if err != nil {
		log.Fatal(msg, zap.Error(err))
	}
}
