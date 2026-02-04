package postgres

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Jitter     float64
}

func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   2 * time.Second,
		Jitter:     0.25,
	}
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if sqlerr, ok := err.(interface{ SQLState() string }); ok {
		if sqlerr.SQLState()[:2] == "40" || sqlerr.SQLState()[:2] == "55" {
			return true
		}
	}

	errStr := err.Error()
	retryableCodes := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"timeout",
		"deadlock",
		"too many connections",
	}
	for _, code := range retryableCodes {
		if contains(errStr, code) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func withRetry(ctx context.Context, config *RetryConfig, operation func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			default:
			}

			delay := config.BaseDelay * time.Duration(1<<uint(attempt-1))
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			jitter := time.Duration(float64(delay) * config.Jitter * (2*rand.Float64() - 1))
			delay = delay + jitter

			log.Warn().
				Int("attempt", attempt).
				Dur("delay", delay).
				Err(lastErr).
				Msg("Retrying operation")

			time.Sleep(delay)
		}

		err := operation()
		if err == nil {
			if attempt > 0 {
				log.Info().
					Int("attempts", attempt+1).
					Msg("Operation succeeded after retry")
			}
			return nil
		}

		lastErr = err

		if !isRetryableError(err) {
			return err
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}
