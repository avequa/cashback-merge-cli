// Package bonus - реализация бонусной части
package bonus

import (
	"cashbackmerge/internal/core/domain"
	"testing"
)

// boardOf собирает квадратное поле из процентов по числу строк, 0 - пустая клетка
func boardOf(t *testing.T, rows [][]uint) domain.Board {
	t.Helper()

	b, err := domain.NewBoard(len(rows))
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
