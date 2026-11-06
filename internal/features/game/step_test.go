package game

import (
	"cashbackmerge/internal/core/apperr"
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/featerr"
	"errors"
	"testing"
)

// fixedRandom всегда выбирает первую свободную клетку, а из Uint64N возвращает x:
// нулевое значение выбирает первое spawn-правило
type fixedRandom struct {
	x uint64
}

func (fixedRandom) IntN(int) int { return 0 }

func (f fixedRandom) Uint64N(uint64) uint64 { return f.x }

// testSize - размер поля в тестах MR2
const testSize = 4

// boardOf собирает поле testSize на testSize из процентов, 0 - пустая клетка,
// недостающие строки и клетки остаются пустыми
func boardOf(t *testing.T, rows [][]uint) domain.Board {
	t.Helper()

	b, err := domain.NewBoard(testSize)
	if err != nil {
		t.Fatal(err)
	}

	for r, row := range rows {
		for c, v := range row {
			if v == 0 {
				continue
			}

			cashback, err := domain.NewCashback(v)
			if err != nil {
				t.Fatal(err)
			}

			if err := b.Place(domain.Tile{Pos: domain.Position{Row: r, Col: c}, Cashback: cashback}); err != nil {
				t.Fatal(err)
			}
		}
	}

	return b
}

func TestApplyMoveWithoutChange(t *testing.T) {
	state := domain.State{
		Board:      boardOf(t, [][]uint{{1, 2, 0, 0}}),
		MovesLimit: 10,
		MovesUsed:  3,
		Score:      5,
	}

	res, err := ApplyMove(state, domain.Left, testCampaign(t).SpawnRules, NewRandom(42))
	if err != nil {
		t.Fatal(err)
	}

	if res.Changed {
		t.Error("Changed = true, want false")
	}

	if res.State != state {
		t.Errorf("state changed: got %+v, want %+v", res.State, state)
	}
}

func TestApplyMoveSpawnsOneTileAndAddsMergeScore(t *testing.T) {
	tests := []struct {
		name       string
		row        []uint
		wantGained int
	}{
		{"two pairs", []uint{2, 2, 4, 4}, 12},
		{"shift without merge", []uint{0, 0, 0, 1}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := domain.State{
				Board:      boardOf(t, [][]uint{tt.row}),
				MovesLimit: 10,
				Score:      5,
			}

			shifted, _, err := state.Board.MergeBoardAfterMove(domain.Left)
			if err != nil {
				t.Fatal(err)
			}

			res, err := ApplyMove(state, domain.Left, testCampaign(t).SpawnRules, NewRandom(42))
			if err != nil {
				t.Fatal(err)
			}

			if !res.Changed {
				t.Fatal("Changed = false, want true")
			}

			if got, want := len(res.State.Board.Tiles()), len(shifted.Tiles())+1; got != want {
				t.Errorf("tiles after move = %d, want %d", got, want)
			}

			if res.Gained != tt.wantGained || res.State.Score != 5+tt.wantGained {
				t.Errorf("Gained = %d, Score = %d, want %d and %d", res.Gained, res.State.Score, tt.wantGained, 5+tt.wantGained)
			}

			if res.State.MovesUsed != 1 {
				t.Errorf("MovesUsed = %d, want 1", res.State.MovesUsed)
			}
		})
	}
}

func TestApplyMoveFinishesAtMovesLimit(t *testing.T) {
	c := testCampaign(t)
	c.MovesLimit = 2

	state, err := NewGame(c, NewRandom(42))
	if err != nil {
		t.Fatal(err)
	}

	rng := NewRandom(7)

	for state.MovesUsed < c.MovesLimit {
		if state.Finished {
			t.Fatalf("finished after %d moves, want %d", state.MovesUsed, c.MovesLimit)
		}

		var moved bool

		for _, d := range domain.Directions() {
			res, err := ApplyMove(state, d, c.SpawnRules, rng)
			if err != nil {
				t.Fatal(err)
			}

			if res.Changed {
				state, moved = res.State, true

				break
			}
		}

		if !moved {
			t.Fatal("no valid move on the board")
		}
	}

	if !state.Finished {
		t.Fatal("Finished = false after movesLimit moves")
	}

	if _, err := ApplyMove(state, domain.Left, c.SpawnRules, rng); !errors.Is(err, featerr.ErrGameFinished) {
		t.Errorf("err = %v, want ErrGameFinished", err)
	}
}

func TestApplyMoveFinishesWhenNoMovesLeft(t *testing.T) {
	// после LEFT в (3, 3) появляется 1% и соседних равных плиток не остаётся
	state := domain.State{
		Board: boardOf(t, [][]uint{
			{2, 4, 2, 4},
			{4, 2, 4, 2},
			{2, 4, 2, 4},
			{0, 8, 16, 8},
		}),
		MovesLimit: 100,
	}

	res, err := ApplyMove(state, domain.Left, testCampaign(t).SpawnRules, fixedRandom{})
	if err != nil {
		t.Fatal(err)
	}

	if !res.State.Finished {
		t.Fatalf("Finished = false, board:\n%s", res.State.Board)
	}

	if res.State.MovesUsed != 1 {
		t.Errorf("MovesUsed = %d, want 1", res.State.MovesUsed)
	}
}

// TestApplyMoveDoesNotFinishAt32 проверяет, что получение плитки 32% не завершает партию:
// игра продолжается до лимита ходов или до отсутствия ходов
func TestApplyMoveDoesNotFinishAt32(t *testing.T) {
	state := domain.State{
		Board:      boardOf(t, [][]uint{{16, 16, 0, 0}}),
		MovesLimit: 100,
	}

	res, err := ApplyMove(state, domain.Left, testCampaign(t).SpawnRules, fixedRandom{})
	if err != nil {
		t.Fatal(err)
	}

	// после слияния 32% стоит в (0, 0), новая плитка 1% появляется в первой свободной клетке (0, 1)
	want := boardOf(t, [][]uint{{32, 1, 0, 0}})
	if res.State.Board != want {
		t.Fatalf("got:\n%s\nwant:\n%s", res.State.Board, want)
	}

	if res.State.Finished {
		t.Error("Finished = true after reaching 32%, want false")
	}

	if res.Gained != 32 || res.State.Score != 32 {
		t.Errorf("Gained = %d, Score = %d, want 32 and 32", res.Gained, res.State.Score)
	}
}

// TestApplyMoveFullBoardWithMergeIsNotFinished проверяет, что заполненное поле
// с соседними равными плитками не завершает партию: слияние ещё возможно
func TestApplyMoveFullBoardWithMergeIsNotFinished(t *testing.T) {
	// после LEFT нижняя строка [8, 16, 1, _], в (3, 3) появляется 1%: поле заполнено, но 1% и 1% можно слить
	state := domain.State{
		Board: boardOf(t, [][]uint{
			{2, 4, 2, 4},
			{4, 2, 4, 2},
			{2, 4, 2, 4},
			{0, 8, 16, 1},
		}),
		MovesLimit: 100,
	}

	res, err := ApplyMove(state, domain.Left, testCampaign(t).SpawnRules, fixedRandom{})
	if err != nil {
		t.Fatal(err)
	}

	if empty := res.State.Board.EmptyCells(); len(empty) != 0 {
		t.Fatalf("board has %d empty cells, want full board:\n%s", len(empty), res.State.Board)
	}

	if res.State.Finished {
		t.Fatalf("Finished = true, but a merge is still possible:\n%s", res.State.Board)
	}
}

// TestApplyMoveInvalidDirection проверяет, что неизвестное направление возвращает
// типизированную ошибку и не выполняет ход
func TestApplyMoveInvalidDirection(t *testing.T) {
	state := domain.State{
		Board:      boardOf(t, [][]uint{{1, 1, 0, 0}}),
		MovesLimit: 10,
	}

	res, err := ApplyMove(state, domain.Direction(0), testCampaign(t).SpawnRules, fixedRandom{})

	var dirErr *apperr.ErrInvalidDirection
	if !errors.As(err, &dirErr) {
		t.Fatalf("err = %v, want *apperr.ErrInvalidDirection", err)
	}

	if res.Changed {
		t.Error("Changed = true, want false")
	}
}
