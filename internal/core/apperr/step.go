// Package apperr содержит типизированные ошибки домена
package apperr

import "fmt"

// ErrInvalidDirection - неизвестное направление хода
type ErrInvalidDirection struct {
	Direction int
}

func (e *ErrInvalidDirection) Error() string {
	return fmt.Sprintf("invalid direction: %d", e.Direction)
}
