package services

import (
	"context"
	"log"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevelProvider interface {
	GetLogLevel() string
}

func NewLogger(ctx context.Context, config LogLevelProvider) (*zap.Logger, error) {
	stdout := zapcore.AddSync(os.Stdout)
	level := getLogLevel(config)
	loggerCfg := zap.NewDevelopmentEncoderConfig()
	if os.Getenv("APP_ENV") == "production" {
		loggerCfg = zap.NewProductionEncoderConfig()
	}

	consoleEncoder := zapcore.NewConsoleEncoder(loggerCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
	)

	logger := zap.New(core)

	return logger, nil
}

func getLogLevel(config LogLevelProvider) zap.AtomicLevel {
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(config.GetLogLevel()))
	if err != nil {
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	return level
}

func SyncLogger(logger *zap.Logger) {
	if err := logger.Sync(); err != nil && !strings.Contains(err.Error(), "sync /dev/stderr: The handle is invalid.") {
		log.Fatalf("can't sync zap logger: %v", err)
	}
}
