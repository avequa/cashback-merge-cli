// Package bonus - реализация бонусной части
package bonus

import (
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/featerr"
	"cashbackmerge/internal/features/game"
	"fmt"
)

// ScoreStrategy - структура бонусной стратегии
type ScoreStrategy struct{}

// Name - возвращает название ScoreStrategy (нужен для соблюдения интерфейса)
func (s ScoreStrategy) Name() string {
	return "greedy-score"
}

// Choose - выбирает лучшее направление
func (s ScoreStrategy) Choose(
	state domain.State,
	_ game.Random,
) (domain.Direction, error) {
	var (
		best              domain.Direction
		bestWin, bestFree int
	)

	for _, direction := range domain.Directions() {
		next, score, err := state.Board.MergeBoardAfterMove(direction)
		if err != nil {
			return 0, fmt.Errorf("invalid move: %d: %w", direction, err)
		}

		if next == state.Board {
			continue
		}

		free := len(next.EmptyCells())
		if best == 0 || score > bestWin || (score == bestWin && free > bestFree) {
			best, bestWin, bestFree = direction, score, free
		}
	}

	if best == 0 {
		return 0, featerr.ErrNoMoves
	}

	return best, nil
}
