package data

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go"
	"go.uber.org/zap"
)

type RetryConfig struct {
	Context          context.Context
	Logger           *zap.Logger
	RecoverableError error
}

func WithRetry(config *RetryConfig, operation func() error) error {
	return retry.Do(
		func() error {
			err := operation()
			if errors.Is(err, config.RecoverableError) {
				return err
			} else if err != nil {
				return retry.Unrecoverable(err)
			}
			return nil
		},
		retry.Context(config.Context),
		retry.Attempts(3),
		retry.OnRetry(func(n uint, err error) {
			config.Logger.Warn(
				"Retrying",
				zap.Uint("RetryAttempt", n),
				zap.Error(err),
			)
		}),
		retry.DelayType(func(n uint, err error, retryConfig *retry.Config) time.Duration {
			return retry.BackOffDelay(n, err, retryConfig)
		}),
		retry.LastErrorOnly(true),
	)
}
