// Package game - сценарии партии: создание начального состояния, появление плиток, ход, стратегии выбора хода
package game

import (
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/featerr"
	"fmt"
)

// MoveResult - результат хода
type MoveResult struct {
	State   domain.State
	Changed bool // ход изменил поле: потрачен ход и появилась новая плитка
	Gained  int  // очки за слияния в этом ходу
}

// ApplyMove выполняет ход: сначала сдвиг и слияние, затем, только если поле изменилось, появление новой плитки.
// Ход без изменения поля возвращает исходное состояние: лимит не расходуется, плитка не появляется
func ApplyMove(state domain.State, d domain.Direction, rules []domain.SpawnRule, rng Random) (MoveResult, error) {
	if state.Finished {
		return MoveResult{}, featerr.ErrGameFinished
	}

	next, gained, err := state.Board.MergeBoardAfterMove(d)
	if err != nil {
		return MoveResult{}, fmt.Errorf("shift cell: %w", err)
	}

	if next == state.Board {
		return MoveResult{State: state}, nil
	}

	// после изменившего поле хода свободная клетка есть всегда: плитки либо слились, либо сдвинулись в пустоту
	if err := spawnTile(&next, rules, rng); err != nil {
		return MoveResult{}, fmt.Errorf("spawn tile: %w", err)
	}

	state.Board = next
	state.MovesUsed++
	state.Score += gained
	state.Finished = state.MovesUsed >= state.MovesLimit || len(availableMoves(state.Board)) == 0

	return MoveResult{State: state, Changed: true, Gained: gained}, nil
}

// availableMoves возвращает направления, ход в которые меняет поле, в порядке domain.Directions()
func availableMoves(b domain.Board) []domain.Direction {
	var moves []domain.Direction

	for _, d := range domain.Directions() {
		next, _, err := b.MergeBoardAfterMove(d)
		if err == nil && next != b {
			moves = append(moves, d)
		}
	}

	return moves
}
