// Package bonus - реализация бонусной части
package bonus

import (
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/featerr"
	"errors"
	"testing"
)

func TestGreedyScoreChoosesMaxScore(t *testing.T) {
	// LEFT/RIGHT сливают 4+4 и 1+1 (+10), UP/DOWN - только 1+1 во втором столбце (+2)
	state := domain.State{Board: boardOf(t, [][]uint{
		{4, 4, 0, 0},
		{0, 1, 0, 0},
		{0, 1, 0, 0},
		{1, 1, 0, 0},
	})}

	got, err := ScoreStrategy{}.Choose(state, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got != domain.Left {
		t.Fatalf("got direction %d, want LEFT", got)
	}
}

func TestGreedyScoreTieBreaksByFreeCells(t *testing.T) {
	// все ходы дают +4, но UP/DOWN сливают 2+2 (одно слияние, 11 свободных клеток),
	// а LEFT/RIGHT - 1+1 и 1+1 (два слияния, 12 свободных клеток)
	state := domain.State{Board: boardOf(t, [][]uint{
		{1, 1, 0, 0},
		{2, 0, 0, 0},
		{0, 0, 1, 1},
		{2, 0, 0, 0},
	})}

	got, err := ScoreStrategy{}.Choose(state, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got != domain.Left {
		t.Fatalf("got direction %d, want LEFT", got)
	}
}

func TestGreedyScoreNoMoves(t *testing.T) {
	state := domain.State{Board: boardOf(t, [][]uint{
		{1, 2, 1, 2},
		{2, 1, 2, 1},
		{1, 2, 1, 2},
		{2, 1, 2, 1},
	})}

	if _, err := (ScoreStrategy{}).Choose(state, nil); !errors.Is(err, featerr.ErrNoMoves) {
		t.Fatalf("got error %v, want ErrNoMoves", err)
	}
}
