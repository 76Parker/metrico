package reporter

import (
	"time"
)

const maxRetries = 3

func withRetry(fn func() error) error {
	retryAttempts := 0
	delay := 1 * time.Second
	for {
		// Если это не первый вызов, то ждем перед повторным вызовом
		if retryAttempts != 0 {
			time.Sleep(delay)
			delay += 2 * time.Second
		}
		err := fn()
		if err == nil {
			return nil
		}
		retryAttempts++
		// Выходим сразу если количество попыток превысило maxRetries
		if retryAttempts > maxRetries || !isRetryable(err) {
			return err
		}
	}
}
