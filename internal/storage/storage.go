package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/salivare/subscriptions-service/internal/config"
)

var (
	ErrSubscriptionExists = errors.New("subscription already exists")
	ErrNotFound           = errors.New("subscription not found")
)

func RetryBackoff(cfg config.RetryConfig, fn func() error) error {
	delay := cfg.InitialDelay

	var err error

	for attempt := 1; attempt <= cfg.Attempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		fmt.Printf(
			"retry failed (attempt %d/%d): %v — retrying in %s\n",
			attempt,
			cfg.Attempts,
			err,
			delay,
		)

		time.Sleep(delay)

		delay += cfg.Step
		if delay > cfg.MaxDelay {
			delay = cfg.InitialDelay
		}
	}

	return fmt.Errorf("all retry attempts failed: %w", err)
}
