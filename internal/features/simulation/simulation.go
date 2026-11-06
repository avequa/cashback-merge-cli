// Package simulation - batch-симуляция кампании: прогон партий, сбор статистики и отчёт
package simulation

import (
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/game"
	"fmt"
	"math/big"
)

// Status - итог проверки ограничения на долю максимальной награды
type Status string

// Статусы проверки: OK - доля в пределах maxRewardAllowedRate, FAILED - превышает
const (
	StatusOK     Status = "OK"
	StatusFailed Status = "FAILED"
)

// Params - параметры запуска симуляции
type Params struct {
	Players  int // больше нуля, проверяется при разборе флагов в cli
	Seed     uint64
	Strategy game.Strategy
}

// Result - статистика по всем партиям симуляции
type Result struct {
	Campaign         domain.Campaign
	Players          int
	Seed             uint64
	Strategy         string
	TileCounts       map[domain.Cashback]int // максимальная плитка -> число игроков
	RewardCounts     map[string]int          // название уровня награды -> число игроков
	TotalScore       int
	MaxScore         int
	MaxRewardPlayers int // игроки с максимальной плиткой не меньше maxRewardCashback
}

// Run играет Players партий стратегией и собирает статистику
func Run(c domain.Campaign, p Params) (Result, error) {
	res := Result{
		Campaign:     c,
		Players:      p.Players,
		Seed:         p.Seed,
		Strategy:     p.Strategy.Name(),
		TileCounts:   make(map[domain.Cashback]int),
		RewardCounts: make(map[string]int),
	}

	seeds := game.NewRandom(p.Seed)

	for i := range p.Players {
		state, err := game.Play(c, p.Strategy, game.NewRandom(seeds.Uint64()))
		if err != nil {
			return Result{}, fmt.Errorf("player %d: %w", i, err)
		}

		res.add(c, state)
	}

	return res, nil
}

// add учитывает итог одной партии
func (r *Result) add(c domain.Campaign, state domain.State) {
	maxTile := state.Board.MaxTile()

	r.TileCounts[maxTile]++
	r.RewardCounts[c.RewardFor(maxTile).Name]++
	r.TotalScore += state.Score
	r.MaxScore = max(r.MaxScore, state.Score)

	if c.IsMaxReward(maxTile) {
		r.MaxRewardPlayers++
	}
}

// MaxRewardRate - фактическая доля игроков с максимальной наградой
func (r Result) MaxRewardRate() *big.Rat {
	return big.NewRat(int64(r.MaxRewardPlayers), int64(r.Players))
}

// AverageScore - средний score по всем партиям
func (r Result) AverageScore() *big.Rat {
	return big.NewRat(int64(r.TotalScore), int64(r.Players))
}

// Status сравнивает фактическую долю максимальной награды с допустимой:
// actualMaxRewardRate <= maxRewardAllowedRate -> OK, иначе FAILED
func (r Result) Status() Status {
	if r.MaxRewardRate().Cmp(r.Campaign.MaxRewardAllowedRate) <= 0 {
		return StatusOK
	}

	return StatusFailed
}
