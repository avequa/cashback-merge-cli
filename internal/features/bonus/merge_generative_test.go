// Package bonus - реализация бонусной части
package bonus

import (
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/game"
	"math/rand/v2"
	"testing"
)

const (
	propertySeed   = 2026
	propertyBoards = 2000
	maxPower       = 7 // плитки от 1% до 2^7 = 128%
)

// values переводит плитку в число процентов: значение Cashback снаружи domain не видно,
// поэтому таблица строится по тем же NewCashback, которыми генерируются поля
func values(t *testing.T) map[domain.Cashback]uint {
	t.Helper()

	m := make(map[domain.Cashback]uint)

	// слияние может удвоить максимальную плитку, поэтому таблица на одну степень больше
	for p := range maxPower + 2 {
		v := uint(1) << p

		c, err := domain.NewCashback(v)
		if err != nil {
			t.Fatal(err)
		}

		m[c] = v
	}

	return m
}

// randomBoard - случайное поле случайного размера, каждая клетка заполнена с вероятностью около 60%
func randomBoard(t *testing.T, rng *rand.Rand) domain.Board {
	t.Helper()

	size := domain.MinBoardSize + rng.IntN(domain.MaxBoardSize-domain.MinBoardSize+1)

	rows := make([][]uint, size)
	for r := range rows {
		rows[r] = make([]uint, size)

		for c := range rows[r] {
			if rng.IntN(10) < 6 {
				rows[r][c] = 1 << rng.IntN(maxPower+1)
			}
		}
	}

	return boardOf(t, rows)
}

// lineValues возвращает значения i-й линии поля в порядке от края, к которому идёт сдвиг d, 0 - пустая клетка
func lineValues(b domain.Board, val map[domain.Cashback]uint, d domain.Direction, i int) []uint {
	n := b.Size()
	cells := make(map[domain.Position]uint)

	for _, tile := range b.Tiles() {
		cells[tile.Pos] = val[tile.Cashback]
	}

	line := make([]uint, n)

	for j := range n {
		var p domain.Position

		switch d {
		case domain.Left:
			p = domain.Position{Row: i, Col: j}
		case domain.Right:
			p = domain.Position{Row: i, Col: n - 1 - j}
		case domain.Up:
			p = domain.Position{Row: j, Col: i}
		case domain.Down:
			p = domain.Position{Row: n - 1 - j, Col: i}
		}

		line[j] = cells[p]
	}

	return line
}

func nonEmpty(line []uint) []uint {
	var out []uint

	for _, v := range line {
		if v != 0 {
			out = append(out, v)
		}
	}

	return out
}

func sum(vals []uint) uint {
	var s uint
	for _, v := range vals {
		s += v
	}

	return s
}

// checkLine проверяет одну линию: каждая плитка результата - это одна исходная плитка
// или ровно две равные соседние исходные плитки, все исходные плитки использованы по одному разу,
// плитки результата прижаты к краю. Возвращает сумму плиток, полученных слиянием
func checkLine(t *testing.T, before, after []uint) uint {
	t.Helper()

	in, out := nonEmpty(before), nonEmpty(after)

	for j, v := range after {
		if (v == 0) != (j >= len(out)) {
			t.Fatalf("line %v -> %v: tiles are not shifted to the edge", before, after)
		}
	}

	var merged uint

	k := 0

	for _, v := range out {
		switch {
		case k < len(in) && in[k] == v:
			k++
		case k+1 < len(in) && in[k] == in[k+1] && in[k]*2 == v:
			merged += v
			k += 2
		default:
			t.Fatalf("line %v -> %v: tile %d is neither a source tile nor a merge of two equal tiles", before, after, v)
		}
	}

	if k != len(in) {
		t.Fatalf("line %v -> %v: source tiles %v are lost", before, after, in[k:])
	}

	return merged
}

func TestMergeGenerative(t *testing.T) {
	val := values(t)
	rng := rand.New(rand.NewPCG(propertySeed, propertySeed))

	for range propertyBoards {
		b := randomBoard(t, rng)

		for _, d := range domain.Directions() {
			next, gained, err := b.MergeBoardAfterMove(d)
			if err != nil {
				t.Fatal(err)
			}

			var sumBefore, sumAfter, merged uint

			for i := range b.Size() {
				before, after := lineValues(b, val, d, i), lineValues(next, val, d, i)

				sumBefore += sum(before)
				sumAfter += sum(after)
				merged += checkLine(t, before, after)
			}

			if sumBefore != sumAfter {
				t.Fatalf("direction %d: tile sum changed %d -> %d\n%s", d, sumBefore, sumAfter, b)
			}

			if uint(gained) != merged {
				t.Fatalf("direction %d: gained %d, want sum of merged tiles %d\n%s", d, gained, merged, b)
			}
		}
	}
}

func TestPlaySameSeedSameGame(t *testing.T) {
	rng := rand.New(rand.NewPCG(propertySeed, propertySeed))

	for range 50 {
		c, err := parseCampaign(t, "")
		if err != nil {
			t.Fatal(err)
		}

		c.BoardSize = domain.MinBoardSize + rng.IntN(domain.MaxBoardSize-domain.MinBoardSize+1)
		seed := rng.Uint64()

		for _, s := range []game.Strategy{game.RandomStrategy{}, ScoreStrategy{}} {
			first, err := game.Play(c, s, game.NewRandom(seed))
			if err != nil {
				t.Fatal(err)
			}

			second, err := game.Play(c, s, game.NewRandom(seed))
			if err != nil {
				t.Fatal(err)
			}

			if first != second {
				t.Fatalf("%s, size %d, seed %d: same seed gave different games", s.Name(), c.BoardSize, seed)
			}
		}
	}
}
