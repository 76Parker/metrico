package app

import (
	"errors"
	"slices"
)

type cleanupStack struct {
	closers []func() error
}

func (s *cleanupStack) add(f func() error) {
	s.closers = append(s.closers, f)
}

func (s *cleanupStack) close() error {
	var err error

	// Освобождаем ресурсы в обратном порядке с помощью slices.Backward
	for _, closer := range slices.Backward(s.closers) {
		err = errors.Join(err, closer())
	}
	s.closers = nil
	return err
}
