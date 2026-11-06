// Package apperr содержит типизированные ошибки домена
package apperr

import "fmt"

// ErrOutOfGameBoard - координаты клетки за пределами поля
type ErrOutOfGameBoard struct {
	Row, Col int
}

func (e *ErrOutOfGameBoard) Error() string {
	return fmt.Sprintf("cell (%d, %d) is out of board bounds", e.Row, e.Col)
}

// ErrCellOccup - в клетке уже стоит плитка
type ErrCellOccup struct {
	Row, Col int
}

func (e *ErrCellOccup) Error() string {
	return fmt.Sprintf("cell (%d, %d) is already occupied", e.Row, e.Col)
}

// ErrInvalidBoardSize - размер поля вне допустимого диапазона
type ErrInvalidBoardSize struct {
	Size, Min, Max int
}

func (e *ErrInvalidBoardSize) Error() string {
	return fmt.Sprintf("board size %d is out of range [%d, %d]", e.Size, e.Min, e.Max)
}
