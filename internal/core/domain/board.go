package domain

import "cashbackmerge/internal/core/apperr"

// Допустимый размер стороны поля, сам размер задаётся в конфигурации кампании
const (
	MinBoardSize = 2
	MaxBoardSize = 8
)

// Position - координаты клетки (отсчёт с нуля от левого верхнего угла)
type Position struct {
	Row, Col int
}

// Tile - плитка с уровнем кэшбэка в клетке поля
type Tile struct {
	Pos      Position
	Cashback Cashback
}

// Board - квадратное поле size на size, пустая клетка хранит Cashback{}.
// Клетки хранятся в массиве максимального размера, используется только угол size на size:
// так поля можно сравнивать через ==
type Board struct {
	size  int
	cells [MaxBoardSize][MaxBoardSize]Cashback
}

// NewBoard создаёт пустое поле size на size
func NewBoard(size int) (Board, error) {
	if size < MinBoardSize || size > MaxBoardSize {
		return Board{}, &apperr.ErrInvalidBoardSize{Size: size, Min: MinBoardSize, Max: MaxBoardSize}
	}

	return Board{size: size}, nil
}

// Size возвращает длину стороны поля
func (b Board) Size() int {
	return b.size
}

// Place ставит плитку в пустую клетку внутри поля
func (b *Board) Place(t Tile) error {
	p := t.Pos
	if p.Row < 0 || p.Row >= b.size || p.Col < 0 || p.Col >= b.size {
		return &apperr.ErrOutOfGameBoard{Row: p.Row, Col: p.Col}
	}

	if b.cells[p.Row][p.Col] != (Cashback{}) {
		return &apperr.ErrCellOccup{Row: p.Row, Col: p.Col}
	}

	b.cells[p.Row][p.Col] = t.Cashback

	return nil
}

// Tiles возвращает плитки поля в порядке обхода по строкам
func (b Board) Tiles() []Tile {
	var tiles []Tile

	for r := range b.size {
		for c := range b.size {
			if v := b.cells[r][c]; v != (Cashback{}) {
				tiles = append(tiles, Tile{Pos: Position{Row: r, Col: c}, Cashback: v})
			}
		}
	}

	return tiles
}

// EmptyCells возвращает свободные клетки в порядке обхода по строкам
func (b Board) EmptyCells() []Position {
	var empty []Position

	for r := range b.size {
		for c := range b.size {
			if b.cells[r][c] == (Cashback{}) {
				empty = append(empty, Position{Row: r, Col: c})
			}
		}
	}

	return empty
}

// MaxTile возвращает максимальную плитку на поле, для пустого поля - Cashback{}
func (b Board) MaxTile() Cashback {
	var m Cashback

	for _, t := range b.Tiles() {
		if t.Cashback.percent > m.percent {
			m = t.Cashback
		}
	}

	return m
}
