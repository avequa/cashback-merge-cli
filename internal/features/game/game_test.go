package game

import (
	"cashbackmerge/internal/core/domain"
	"math/big"
	"testing"
)

func testCampaign(t *testing.T) domain.Campaign {
	t.Helper()

	cashback := func(s string) domain.Cashback {
		c, err := domain.ParseCashback(s)
		if err != nil {
			t.Fatal(err)
		}

		return c
	}

	return domain.Campaign{
		BoardSize:  4,
		MovesLimit: 100,
		SpawnRules: []domain.SpawnRule{
			{Cashback: cashback("1"), Probability: big.NewRat(85, 100)},
			{Cashback: cashback("2"), Probability: big.NewRat(15, 100)},
		},
		MaxRewardCashback:    cashback("32"),
		MaxRewardAllowedRate: big.NewRat(20, 100),
		Rewards: []domain.Reward{
			{MinTile: cashback("1"), Name: "basic"},
			{MinTile: cashback("32"), Name: "platinum"},
		},
	}
}

func TestNewGameHasTwoTilesInDifferentCells(t *testing.T) {
	state, err := NewGame(testCampaign(t), NewRandom(42))
	if err != nil {
		t.Fatal(err)
	}

	tiles := state.Board.Tiles()
	if len(tiles) != 2 {
		t.Fatalf("got %d tiles, want 2", len(tiles))
	}

	if tiles[0].Pos == tiles[1].Pos {
		t.Fatalf("both tiles are in cell %v", tiles[0].Pos)
	}
}

func TestNewGameSameSeedSameBoard(t *testing.T) {
	first, err := NewGame(testCampaign(t), NewRandom(42))
	if err != nil {
		t.Fatal(err)
	}

	second, err := NewGame(testCampaign(t), NewRandom(42))
	if err != nil {
		t.Fatal(err)
	}

	if first.Board != second.Board {
		t.Fatalf("same seed gave different boards:\n%s\n%s", first.Board, second.Board)
	}
}

// TestPickCashbackIntervals проверяет границы интервалов взвешенного выбора без статистики:
// 0.85 и 0.15 приводятся к 17/20 и 3/20, поэтому x в [0, 17) даёт 1%, а x в [17, 20) - 2%
func TestPickCashbackIntervals(t *testing.T) {
	rules := testCampaign(t).SpawnRules

	tests := []struct {
		x    uint64
		want domain.Cashback
	}{
		{0, rules[0].Cashback},
		{16, rules[0].Cashback},
		{17, rules[1].Cashback},
		{19, rules[1].Cashback},
	}

	for _, tt := range tests {
		got, err := pickCashback(rules, fixedRandom{x: tt.x})
		if err != nil {
			t.Fatal(err)
		}

		if got != tt.want {
			t.Errorf("x = %d: got %s, want %s", tt.x, got, tt.want)
		}
	}
}
