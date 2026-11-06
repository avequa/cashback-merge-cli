// Package domain - содержит описание доменных сущностей: игровой доски, кампании, кэшбека, состояния партии
package domain

import "cashbackmerge/internal/core/apperr"

// Direction - задаёт направление движения по игровой доске
type Direction int

// Направления хода, нулевое значение не является направлением
const (
	Up    Direction = 1
	Down  Direction = 2
	Left  Direction = 3
	Right Direction = 4
)

// Directions возвращает все направления хода
func Directions() []Direction {
	return []Direction{Up, Down, Left, Right}
}

// MergeBoardAfterMove сдвигает все плитки к краю d и сливает соседние равные
func (b Board) MergeBoardAfterMove(d Direction) (Board, int, error) {
	if d < Up || d > Right {
		return Board{}, 0, &apperr.ErrInvalidDirection{Direction: int(d)}
	}

	next := Board{size: b.size}

	have := 0

	for i := range b.size {
		positions := rulesReadCells(d, i, b.size)

		line := make([]Cashback, len(positions))
		for j, p := range positions {
			line[j] = b.cells[p.Row][p.Col]
		}

		merged, score := mergeLineAfterShift(line)
		have += score

		for j, p := range positions {
			next.cells[p.Row][p.Col] = merged[j]
		}
	}

	return next, have, nil
}

// rulesReadCells возвращает клетки i-й линии в порядке от края, к которому идёт сдвиг:
// так любое направление сводится к сдвигу строки влево
func rulesReadCells(d Direction, i, size int) []Position {
	line := make([]Position, size)

	for j := range size {
		switch d {
		case Left:
			line[j] = Position{Row: i, Col: j}
		case Right:
			line[j] = Position{Row: i, Col: size - 1 - j}
		case Up:
			line[j] = Position{Row: j, Col: i}
		case Down:
			line[j] = Position{Row: size - 1 - j, Col: i}
		}
	}

	return line
}

// mergeLineAfterShift сдвигает плитки строки к началу и сливает соседние равные слева направо
// Плитка, полученная слиянием, сразу записывается в результат и дальше не сравнивается,
// поэтому одна плитка не сливается дважды за ход: [1, 1, 2, .] -> [2, 2, ., .]
func mergeLineAfterShift(line []Cashback) ([]Cashback, int) {
	tiles := make([]Cashback, 0, len(line))

	for _, v := range line {
		if v != (Cashback{}) {
			tiles = append(tiles, v)
		}
	}

	out := make([]Cashback, len(line))

	have := 0
	n := 0

	for i := 0; i < len(tiles); i++ {
		v := tiles[i]

		if i+1 < len(tiles) && tiles[i+1] == v {
			v = Cashback{percent: v.percent * 2}
			have += int(v.percent)
			i++
		}

		out[n] = v
		n++
	}

	return out, have
}
