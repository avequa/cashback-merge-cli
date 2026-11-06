// Package game - сценарии партии: создание начального состояния, появление плиток, ход, стратегии выбора хода
package game

import (
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/featerr"
)

// Strategy выбирает направление следующего хода.
// Новая стратегия реализует этот интерфейс и добавляется в cli-слое
type Strategy interface {
	Name() string
	Choose(state domain.State, rng Random) (domain.Direction, error)
}

// RandomStrategy равновероятно выбирает одно из направлений, которые меняют поле
type RandomStrategy struct{}

// Name возвращает название стратегии для командной строки и отчёта
func (RandomStrategy) Name() string {
	return "random"
}

// Choose выбирает случайное направление из доступных
func (RandomStrategy) Choose(state domain.State, rng Random) (domain.Direction, error) {
	moves := availableMoves(state.Board)
	if len(moves) == 0 {
		return 0, featerr.ErrNoMoves
	}

	return moves[rng.IntN(len(moves))], nil
}
