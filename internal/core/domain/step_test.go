package domain

import (
	"cashbackmerge/internal/core/apperr"
	"errors"
	"slices"
	"testing"
)

// testSize - размер поля в тестах MR1-MR2
const testSize = 4

// line собирает строку длины testSize из процентов, 0 - пустая клетка
func line(vals ...uint) []Cashback {
	l := make([]Cashback, testSize)
	for i, v := range vals {
		l[i] = Cashback{percent: v}
	}

	return l
}

// boardOf собирает поле testSize на testSize из процентов, 0 - пустая клетка,
// недостающие строки и клетки остаются пустыми
func boardOf(rows [][]uint) Board {
	b := Board{size: testSize}

	for r, row := range rows {
		for c, v := range row {
			b.cells[r][c] = Cashback{percent: v}
		}
	}

	return b
}

// orient переносит поле, записанное для сдвига влево, в направление d:
// клетка, стоящая в строке на месте c от левого края, встаёт на место c от края d
func orient(b Board, d Direction) Board {
	out := Board{size: b.size}

	for r := range b.size {
		for c := range b.size {
			v := b.cells[r][c]

			switch d {
			case Left:
				out.cells[r][c] = v
			case Right:
				out.cells[r][b.size-1-c] = v
			case Up:
				out.cells[c][r] = v
			case Down:
				out.cells[b.size-1-c][r] = v
			}
		}
	}

	return out
}

func Test(t *testing.T) {
	tests := []struct {
		name       string
		in, want   []Cashback
		wantGained int
	}{
		{"pair", line(1, 1, 0, 0), line(2, 0, 0, 0), 2},
		{"three equal", line(1, 1, 1, 0), line(2, 1, 0, 0), 2},
		{"four equal", line(1, 1, 1, 1), line(2, 2, 0, 0), 4},
		{"two pairs", line(2, 2, 4, 4), line(4, 8, 0, 0), 12},
		{"no double merge", line(1, 1, 2, 0), line(2, 2, 0, 0), 2},
		{"merge through gap", line(2, 0, 2, 0), line(4, 0, 0, 0), 4},
		{"shift without merge", line(0, 0, 0, 1), line(1, 0, 0, 0), 0},
		{"empty", line(), line(), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gained := mergeLineAfterShift(tt.in)

			if !slices.Equal(got, tt.want) {
				t.Errorf("mergeLineAfterShift(%v) = %v, want %v", tt.in, got, tt.want)
			}

			if gained != tt.wantGained {
				t.Errorf("gained = %d, want %d", gained, tt.wantGained)
			}
		})
	}
}

func TestShiftAllDirectionsSymmetric(t *testing.T) {
	start := boardOf([][]uint{
		{1, 1, 2, 0},
		{2, 2, 4, 4},
		{0, 1, 0, 1},
		{4, 0, 0, 4},
	})

	wantLeft := boardOf([][]uint{
		{2, 2, 0, 0},
		{4, 8, 0, 0},
		{2, 0, 0, 0},
		{8, 0, 0, 0},
	})

	directions := []struct {
		name string
		d    Direction
	}{
		{"up", Up},
		{"down", Down},
		{"left", Left},
		{"right", Right},
	}

	for _, tt := range directions {
		t.Run(tt.name, func(t *testing.T) {
			got, gained, err := orient(start, tt.d).MergeBoardAfterMove(tt.d)
			if err != nil {
				t.Fatal(err)
			}

			if want := orient(wantLeft, tt.d); got != want {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}

			if gained != 24 {
				t.Errorf("gained = %d, want 24", gained)
			}
		})
	}
}

func TestShiftWithoutChange(t *testing.T) {
	b := boardOf([][]uint{
		{1, 2, 0, 0},
		{4, 0, 0, 0},
	})

	got, gained, err := b.MergeBoardAfterMove(Left)
	if err != nil {
		t.Fatal(err)
	}

	if got != b || gained != 0 {
		t.Errorf("board changed or gained = %d:\n%s", gained, got)
	}
}

func TestShiftInvalidDirection(t *testing.T) {
	_, _, err := Board{}.MergeBoardAfterMove(Direction(0))

	var dirErr *apperr.ErrInvalidDirection
	if !errors.As(err, &dirErr) {
		t.Fatalf("err = %v, want *apperr.ErrInvalidDirection", err)
	}
}
