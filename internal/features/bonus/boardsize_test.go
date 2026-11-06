// Package bonus - реализация бонусной части
package bonus

import (
	"cashbackmerge/internal/config"
	"cashbackmerge/internal/core/apperr"
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/game"
	"errors"
	"strings"
	"testing"
)

// campaignJSON - конфигурация кампании, boardSize подставляется в начало объекта
const campaignJSON = `{
  %s
  "movesLimit": 100,
  "spawnRules": [{"cashback": "1", "probability": "0.85"}, {"cashback": "2", "probability": "0.15"}],
  "maxRewardCashback": "32",
  "maxRewardAllowedRate": "0.20",
  "rewards": [{"minTile": "1", "name": "basic"}, {"minTile": "32", "name": "platinum"}]
}`

func parseCampaign(t *testing.T, boardSize string) (domain.Campaign, error) {
	t.Helper()

	return config.Parse(strings.NewReader(strings.Replace(campaignJSON, "%s", boardSize, 1)))
}

func TestBoardSizeFromConfig(t *testing.T) {
	tests := []struct {
		name, field string
		want        int
	}{
		{"default without boardSize", ``, 4},
		{"explicit 4", `"boardSize": 4,`, 4},
		{"explicit 5", `"boardSize": 5,`, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := parseCampaign(t, tt.field)
			if err != nil {
				t.Fatal(err)
			}

			state, err := game.NewGame(c, game.NewRandom(42))
			if err != nil {
				t.Fatal(err)
			}

			if got := state.Board.Size(); got != tt.want {
				t.Errorf("board size = %d, want %d", got, tt.want)
			}

			if n := len(state.Board.Tiles()) + len(state.Board.EmptyCells()); n != tt.want*tt.want {
				t.Errorf("board has %d cells, want %d", n, tt.want*tt.want)
			}
		})
	}
}

func TestBoardSizeInvalidConfig(t *testing.T) {
	for _, field := range []string{`"boardSize": 0,`, `"boardSize": 1,`, `"boardSize": 9,`} {
		_, err := parseCampaign(t, field)

		var campaignErr *apperr.ErrInvalidCampaign
		if !errors.As(err, &campaignErr) || campaignErr.Field != "boardSize" {
			t.Errorf("%s: err = %v, want invalid boardSize", field, err)
		}
	}
}

func TestNewBoardInvalidSize(t *testing.T) {
	for _, size := range []int{0, 1, 9} {
		var sizeErr *apperr.ErrInvalidBoardSize
		if _, err := domain.NewBoard(size); !errors.As(err, &sizeErr) {
			t.Errorf("NewBoard(%d): err = %v, want *apperr.ErrInvalidBoardSize", size, err)
		}
	}
}

func TestMergeOnBoard5x5(t *testing.T) {
	start := boardOf(t, [][]uint{
		{1, 1, 1, 1, 1},
		{2, 2, 4, 4, 8},
		{0, 0, 0, 0, 1},
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
	})
	want := boardOf(t, [][]uint{
		{2, 2, 1, 0, 0},
		{4, 8, 8, 0, 0},
		{1, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
	})

	got, gained, err := start.MergeBoardAfterMove(domain.Left)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}

	// 2 + 2 в первой строке, 4 + 8 во второй
	if gained != 16 {
		t.Errorf("gained = %d, want 16", gained)
	}
}

func TestPlayOnBoard5x5(t *testing.T) {
	c, err := parseCampaign(t, `"boardSize": 5,`)
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []game.Strategy{game.RandomStrategy{}, ScoreStrategy{}} {
		state, err := game.Play(c, s, game.NewRandom(42))
		if err != nil {
			t.Fatalf("%s: %v", s.Name(), err)
		}

		if !state.Finished || state.Board.Size() != 5 {
			t.Errorf("%s: finished = %t, size = %d", s.Name(), state.Finished, state.Board.Size())
		}
	}
}
