// Package apperr содержит типизированные ошибки домена
package apperr

import "fmt"

// ErrInvalidCashback - если значение не может быть уровнем кэшбэка
type ErrInvalidCashback[T any] struct {
	Input T
}

func (e *ErrInvalidCashback[T]) Error() string {
	return fmt.Sprintf("invalid cashback %v: must be a positive power of two", e.Input)
}
