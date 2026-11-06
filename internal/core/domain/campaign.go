// Package domain - содержит описание доменных сущностей: игровой доски, кампании, кэшбека, состояния партии
package domain

import (
	"cashbackmerge/internal/core/apperr"
	"fmt"
	"math/big"
)

// SpawnRule - значение новой плитки и вероятность его появления
type SpawnRule struct {
	Cashback    Cashback
	Probability *big.Rat
}

// Reward - уровень награды: игрок получает уровень с наибольшим MinTile, который не больше его максимальной плитки
type Reward struct {
	MinTile Cashback
	Name    string
}

// Campaign - параметры кампании и таблица наград
type Campaign struct {
	BoardSize            int
	MovesLimit           int
	SpawnRules           []SpawnRule
	MaxRewardCashback    Cashback
	MaxRewardAllowedRate *big.Rat
	Rewards              []Reward
}

// Validate проверяет все параметры кампании
func (c Campaign) Validate() error {
	if c.BoardSize < MinBoardSize || c.BoardSize > MaxBoardSize {
		return invalid("boardSize", fmt.Sprintf("must be in [%d, %d]", MinBoardSize, MaxBoardSize))
	}

	if c.MovesLimit <= 0 {
		return invalid("movesLimit", "must be positive")
	}

	if err := c.validateSpawnRules(); err != nil {
		return err
	}

	rate := c.MaxRewardAllowedRate
	if rate == nil || rate.Sign() < 0 || rate.Cmp(big.NewRat(1, 1)) > 0 {
		return invalid("maxRewardAllowedRate", "must be in [0, 1]")
	}

	return c.validateRewards()
}

// validateSpawnRules: правила не пусты, каждая вероятность в (0, 1]
func (c Campaign) validateSpawnRules() error {
	if len(c.SpawnRules) == 0 {
		return invalid("spawnRules", "must not be empty")
	}

	one := big.NewRat(1, 1)
	sum := new(big.Rat)

	for i, r := range c.SpawnRules {
		p := r.Probability
		if p == nil || p.Sign() <= 0 || p.Cmp(one) > 0 {
			return invalid(fmt.Sprintf("spawnRules[%d].probability", i), "must be in (0, 1]")
		}

		sum.Add(sum, p)
	}

	if sum.Cmp(one) != 0 {
		return invalid("spawnRules", "probabilities must sum to 1, got "+sum.RatString())
	}

	return nil
}

// validateRewards проверяет, что таблица полная и уровень по максимальной плитке определяется однозначно:
//   - MinTile строго возрастает, поэтому подходящий уровень ровно один
//   - нижний уровень не выше минимальной spawn-плитки, поэтому уровень есть у любой партии
//   - верхний уровень равен maxRewardCashback, поэтому он и есть максимальная награда
func (c Campaign) validateRewards() error {
	if len(c.Rewards) == 0 {
		return invalid("rewards", "must not be empty")
	}

	for i, r := range c.Rewards {
		if r.Name == "" {
			return invalid(fmt.Sprintf("rewards[%d].name", i), "must not be empty")
		}

		if i > 0 && r.MinTile.percent <= c.Rewards[i-1].MinTile.percent {
			return invalid(fmt.Sprintf("rewards[%d].minTile", i), "must be greater than previous level")
		}
	}

	if lowest := c.Rewards[0].MinTile; lowest.percent > c.minSpawn().percent {
		return invalid("rewards", "lowest level "+lowest.String()+" is above the smallest spawn tile")
	}

	if highest := c.Rewards[len(c.Rewards)-1].MinTile; highest != c.MaxRewardCashback {
		return invalid("rewards", "highest level "+highest.String()+" must equal maxRewardCashback")
	}

	return nil
}

// RewardFor возвращает уровень награды по максимальной плитке:
// уровень с наибольшим MinTile, который не больше maxTile
func (c Campaign) RewardFor(maxTile Cashback) Reward {
	reward := c.Rewards[0]

	for _, r := range c.Rewards[1:] {
		if r.MinTile.percent > maxTile.percent {
			break
		}

		reward = r
	}

	return reward
}

// IsMaxReward сообщает, получил ли игрок с такой максимальной плиткой максимальную награду
func (c Campaign) IsMaxReward(maxTile Cashback) bool {
	return maxTile.percent >= c.MaxRewardCashback.percent
}

// minSpawn возвращает минимальное значение из spawn-правил
func (c Campaign) minSpawn() Cashback {
	m := c.SpawnRules[0].Cashback
	for _, r := range c.SpawnRules[1:] {
		if r.Cashback.percent < m.percent {
			m = r.Cashback
		}
	}

	return m
}

func invalid(field, reason string) error {
	return &apperr.ErrInvalidCampaign{Field: field, Reason: reason}
}
