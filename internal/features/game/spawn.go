// Package game - сценарии партии: создание начального состояния, появление плиток
package game

import (
	"cashbackmerge/internal/core/domain"
	"errors"
	"fmt"
	"math/big"
)

// spawnTile ставит новую плитку в случайную свободную клетку, значение выбирается по вероятностям из spawnRules
func spawnTile(b *domain.Board, rules []domain.SpawnRule, rng Random) error {
	empty := b.EmptyCells()
	if len(empty) == 0 {
		return errors.New("no empty cells")
	}

	cashback, err := pickCashback(rules, rng)
	if err != nil {
		return err
	}

	if err := b.Place(domain.Tile{Pos: empty[rng.IntN(len(empty))], Cashback: cashback}); err != nil {
		return fmt.Errorf("place tile: %w", err)
	}

	return nil
}

// pickCashback выбирает значение по вероятностям:
// дроби приводятся к общему знаменателю и выбор идёт по целым весам
func pickCashback(rules []domain.SpawnRule, rng Random) (domain.Cashback, error) {
	denom := big.NewInt(1)
	for _, r := range rules {
		denom = lcm(denom, r.Probability.Denom())
	}

	if !denom.IsUint64() {
		return domain.Cashback{}, errors.New("spawn probabilities are too precise")
	}

	x := new(big.Int).SetUint64(rng.Uint64N(denom.Uint64()))
	acc := new(big.Int)

	for _, r := range rules {
		weight := new(big.Int).Div(denom, r.Probability.Denom())
		weight.Mul(weight, r.Probability.Num())
		acc.Add(acc, weight)

		if x.Cmp(acc) < 0 {
			return r.Cashback, nil
		}
	}

	// недостижимо: кампания проверена, сумма вероятностей равна 1
	return rules[len(rules)-1].Cashback, nil
}

func lcm(a, b *big.Int) *big.Int {
	gcd := new(big.Int).GCD(nil, nil, a, b)

	return new(big.Int).Mul(a, new(big.Int).Div(b, gcd))
}
