package infrastucture

import (
	"context"
	"log"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(ctx context.Context, config *Config) (*zap.Logger, error) {
	stdout := zapcore.AddSync(os.Stdout)
	stderr := zapcore.AddSync(os.Stderr)
	level := getLogLevel(config)
	loggerCfg := zap.NewDevelopmentEncoderConfig()
	if os.Getenv("APP_ENV") == "production" {
		loggerCfg = zap.NewProductionEncoderConfig()
	}

	consoleEncoder := zapcore.NewConsoleEncoder(loggerCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(consoleEncoder, stderr, level),
	)

	logger := zap.New(core)

	defer func(logger *zap.Logger) {
		// Error is written if OS didn't take care of flushing buffers out
		if err := logger.Sync(); err != nil && !strings.Contains(err.Error(), "sync /dev/stderr: The handle is invalid.") {
			log.Fatalf("can't sync zap logger: %v", err)
		}
	}(logger)

	return logger, nil
}

func getLogLevel(config *Config) zap.AtomicLevel {
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(config.LogLevel))
	if err != nil {
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	return level
}
