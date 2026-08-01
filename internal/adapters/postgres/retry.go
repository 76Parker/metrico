package postgres

import "time"

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
		// Выходим сразу если функция не вернула ошибки
		if err == nil {
			return nil
		}
		retryAttempts++
		// Выходим сразу если количество попыток превысило maxRetries ИЛИ если ошибка не является retryable
		if retryAttempts > maxRetries || !isRetryable(err) {
			return err
		}
	}
}
