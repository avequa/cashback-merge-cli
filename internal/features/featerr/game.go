// Package featerr содержит ошибки features
package featerr

import "errors"

var (
	// ErrGameFinished - ход в уже завершённой партии
	ErrGameFinished = errors.New("game is finished")
	// ErrNoMoves - на поле нет хода, который его изменит
	ErrNoMoves = errors.New("no moves change the board")
	// ErrUnknownStrategy - стратегии выбора хода с таким названием нет
	ErrUnknownStrategy = errors.New("unknown strategy")
	// ErrRecordMismatch - ход из записи партии не меняет поле: запись не от этой кампании или seed
	ErrRecordMismatch = errors.New("move does not change the board, record does not match campaign or seed")
)
