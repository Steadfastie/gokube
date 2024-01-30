package data

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go"
	"go.uber.org/zap"
)

type RetryConfig struct {
	Context           context.Context
	Logger            *zap.Logger
	RecoverableErrors []error
}

func WithRetry(config *RetryConfig, operation func() error) error {
	return retry.Do(
		func() error {
			err := operation()
			if containsError(config.RecoverableErrors, err) {
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

func containsError(recoverableErrors []error, target error) bool {
	for _, err := range recoverableErrors {
		if errors.Is(target, err) {
			return true
		}
	}
	return false
}
