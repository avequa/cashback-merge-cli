// Package bonus - реализация бонусной части
package bonus

import (
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/game"
	"strings"
	"testing"
)

func TestParseDirection(t *testing.T) {
	tests := []struct {
		in   string
		want domain.Direction
		ok   bool
	}{
		{"w", domain.Up, true},
		{"UP", domain.Up, true},
		{"s", domain.Down, true},
		{" down ", domain.Down, true},
		{"a", domain.Left, true},
		{"Left", domain.Left, true},
		{"d", domain.Right, true},
		{"right", domain.Right, true},
		{"x", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		got, ok := ParseDirection(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Errorf("ParseDirection(%q) = %d, %t, want %d, %t", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

func TestPlayUntilMovesLimit(t *testing.T) {
	c, err := parseCampaign(t, "")
	if err != nil {
		t.Fatal(err)
	}

	c.MovesLimit = 2

	// хотя бы одно из четырёх направлений меняет поле, поэтому двух кругов хватит на два хода
	var out strings.Builder

	state, moves, err := Play(c, 42, strings.NewReader(strings.Repeat("w\ns\na\nd\n", 2)), &out)
	if err != nil {
		t.Fatal(err)
	}

	if !state.Finished || state.MovesUsed != 2 || len(moves) != 2 {
		t.Fatalf("finished = %t, moves used = %d, recorded moves = %d", state.Finished, state.MovesUsed, len(moves))
	}

	for _, want := range []string{"Moves: 0/2  Left: 2", "Moves: 2/2  Left: 0", "Reward: ", "Game over"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output does not contain %q:\n%s", want, out.String())
		}
	}
}

func TestPlayMoveWithoutChange(t *testing.T) {
	c, err := parseCampaign(t, "")
	if err != nil {
		t.Fatal(err)
	}

	start, err := game.NewGame(c, game.NewRandom(42))
	if err != nil {
		t.Fatal(err)
	}

	// ищем на начальном поле направление, которое его не меняет
	var cmd string

	for name, d := range directionNames {
		if next, _, mergeErr := start.Board.MergeBoardAfterMove(d); mergeErr == nil && next == start.Board {
			cmd = name
		}
	}

	if cmd == "" {
		t.Fatalf("every direction changes the start board:\n%s", start.Board)
	}

	var out strings.Builder

	state, moves, err := Play(c, 42, strings.NewReader(cmd+"\nq\n"), &out)
	if err != nil {
		t.Fatal(err)
	}

	if state != start || len(moves) != 0 {
		t.Fatalf("move %q without change modified the game: moves used = %d", cmd, state.MovesUsed)
	}

	if !strings.Contains(out.String(), "move does not change the board") {
		t.Errorf("output does not report move without change:\n%s", out.String())
	}
}

func TestPlayUnknownCommandAndQuit(t *testing.T) {
	c, err := parseCampaign(t, "")
	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder

	state, moves, err := Play(c, 42, strings.NewReader("x\nq\nw\n"), &out)
	if err != nil {
		t.Fatal(err)
	}

	if state.Finished || state.MovesUsed != 0 || len(moves) != 0 {
		t.Fatalf("game changed after unknown command and quit: %+v", state)
	}

	if !strings.Contains(out.String(), `unknown command "x"`) {
		t.Errorf("output does not report unknown command:\n%s", out.String())
	}
}
