package random

import (
	"errors"
	"math/rand/v2"
)

type StdRandomGenerator[T any] struct{}

func NewStdRandomGenerator[T any]() *StdRandomGenerator[T] {
	return &StdRandomGenerator[T]{}
}

func (s *StdRandomGenerator[T]) IntN(n int) int {
	return rand.IntN(n)
}

func (s *StdRandomGenerator[T]) PickOne(items []T) (T, error) {
	var zero T
	if len(items) == 0 {
		return zero, errors.New("slice cannot be empty")
	}
	idx := rand.IntN(len(items))
	return items[idx], nil
}
