// Package game - сценарии партии: создание начального состояния, появление плиток, ход, стратегии выбора хода
package game

import (
	"cashbackmerge/internal/core/domain"
	"fmt"
	"math/rand/v2"
)

// initialTiles - число плиток на начальном поле
const initialTiles = 2

// Random - источник случайности
type Random interface {
	IntN(n int) int
	Uint64N(n uint64) uint64
}

// NewRandom создаёт локальный генератор (одинаковый seed даёт одинаковую последовательность)
func NewRandom(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed)) //nolint:gosec
}

// NewGame создаёт начальное состояние: пустое поле размера из кампании и две плитки в разных клетках
func NewGame(c domain.Campaign, rng Random) (domain.State, error) {
	if err := c.Validate(); err != nil {
		return domain.State{}, fmt.Errorf("validate campaign: %w", err)
	}

	board, err := domain.NewBoard(c.BoardSize)
	if err != nil {
		return domain.State{}, fmt.Errorf("new board: %w", err)
	}

	state := domain.State{Board: board, MovesLimit: c.MovesLimit}

	for range initialTiles {
		if err := spawnTile(&state.Board, c.SpawnRules, rng); err != nil {
			return domain.State{}, fmt.Errorf("spawn initial tile: %w", err)
		}
	}

	return state, nil
}

// Play играет партию стратегией от начального поля до завершения.
// Одинаковые кампания, стратегия и состояние rng дают одинаковую партию
func Play(c domain.Campaign, s Strategy, rng Random) (domain.State, error) {
	state, err := NewGame(c, rng)
	if err != nil {
		return domain.State{}, err
	}

	for !state.Finished {
		d, err := s.Choose(state, rng)
		if err != nil {
			return domain.State{}, fmt.Errorf("strategy %s: %w", s.Name(), err)
		}

		res, err := ApplyMove(state, d, c.SpawnRules, rng)
		if err != nil {
			return domain.State{}, fmt.Errorf("apply move: %w", err)
		}

		// иначе стратегия, выбирающая ход без изменения поля, зациклит партию
		if !res.Changed {
			return domain.State{}, fmt.Errorf("strategy %s: move %d does not change the board", s.Name(), d)
		}

		state = res.State
	}

	return state, nil
}
