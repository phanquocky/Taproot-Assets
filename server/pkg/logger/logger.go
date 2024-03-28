package logger

import (
	"log"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	singleton *zap.Logger
	once      sync.Once
)

func Init() {
	zaplv, err := zapcore.ParseLevel("info")
	if err != nil {
		log.Fatalf("Cannot init logger")
	}

	zapcfg := zap.Config{
		Encoding:    "console",
		Level:       zap.NewAtomicLevelAt(zaplv),
		OutputPaths: []string{"stderr"},

		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:    "message",
			StacktraceKey: "stacktrace",
			TimeKey:       "time",
			LevelKey:      "level",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey,
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeCaller:  zapcore.FullCallerEncoder,
			EncodeTime:    zapcore.RFC3339TimeEncoder,
		},
	}

	logger, err := zapcfg.Build()
	if err != nil {
		log.Fatalf("Cannot init logger")
	}

	once.Do(func() {
		singleton = logger
	})
}

func Infow(message string, keysAndValues ...any) {
	singleton.Sugar().Infow(message, keysAndValues...)
}

func Errorw(message string, keysAndValues ...any) {
	singleton.Sugar().Errorw(message, keysAndValues...)
}
