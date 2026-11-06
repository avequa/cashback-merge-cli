// Package config - загрузка кампании из JSON-файла в domain/campaign
package config

import (
	"bytes"
	"cashbackmerge/internal/core/apperr"
	"cashbackmerge/internal/core/domain"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
)

// defaultBoardSize - размер поля, если boardSize не указан в конфигурации
const defaultBoardSize = 4

// campaignJSON - формат файла конфигурации
type campaignJSON struct {
	BoardSize            *int            `json:"boardSize"`
	MovesLimit           int             `json:"movesLimit"`
	SpawnRules           []spawnRuleJSON `json:"spawnRules"`
	MaxRewardCashback    string          `json:"maxRewardCashback"`
	MaxRewardAllowedRate string          `json:"maxRewardAllowedRate"`
	Rewards              []rewardJSON    `json:"rewards"`
}

type spawnRuleJSON struct {
	Cashback    string `json:"cashback"`
	Probability string `json:"probability"`
}

type rewardJSON struct {
	MinTile string `json:"minTile"`
	Name    string `json:"name"`
}

// Load читает кампанию из файла и проверяет её
func Load(path string) (domain.Campaign, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return domain.Campaign{}, fmt.Errorf("read config: %w", err)
	}

	return Parse(bytes.NewReader(data))
}

// Parse читает кампанию из r и проверяет её, неизвестные поля считаются ошибкой
func Parse(r io.Reader) (domain.Campaign, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var raw campaignJSON
	if err := dec.Decode(&raw); err != nil {
		return domain.Campaign{}, fmt.Errorf("decode config: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return domain.Campaign{}, errors.New("decode config: unexpected data after JSON object")
	}

	c, err := raw.toDomain()
	if err != nil {
		return domain.Campaign{}, err
	}

	if err := c.Validate(); err != nil {
		return domain.Campaign{}, fmt.Errorf("validate config: %w", err)
	}

	return c, nil
}

// toDomain переводит строки из JSON в типы домена
func (raw campaignJSON) toDomain() (domain.Campaign, error) {
	spawnRules := make([]domain.SpawnRule, 0, len(raw.SpawnRules))

	for i, r := range raw.SpawnRules {
		cashback, err := domain.ParseCashback(r.Cashback)
		if err != nil {
			return domain.Campaign{}, fmt.Errorf("spawnRules[%d].cashback: %w", i, err)
		}

		probability, err := parseRat(fmt.Sprintf("spawnRules[%d].probability", i), r.Probability)
		if err != nil {
			return domain.Campaign{}, err
		}

		spawnRules = append(spawnRules, domain.SpawnRule{Cashback: cashback, Probability: probability})
	}

	maxReward, err := domain.ParseCashback(raw.MaxRewardCashback)
	if err != nil {
		return domain.Campaign{}, fmt.Errorf("maxRewardCashback: %w", err)
	}

	rate, err := parseRat("maxRewardAllowedRate", raw.MaxRewardAllowedRate)
	if err != nil {
		return domain.Campaign{}, err
	}

	rewards := make([]domain.Reward, 0, len(raw.Rewards))

	for i, r := range raw.Rewards {
		minTile, err := domain.ParseCashback(r.MinTile)
		if err != nil {
			return domain.Campaign{}, fmt.Errorf("rewards[%d].minTile: %w", i, err)
		}

		rewards = append(rewards, domain.Reward{MinTile: minTile, Name: r.Name})
	}

	boardSize := defaultBoardSize
	if raw.BoardSize != nil {
		boardSize = *raw.BoardSize
	}

	return domain.Campaign{
		BoardSize:            boardSize,
		MovesLimit:           raw.MovesLimit,
		SpawnRules:           spawnRules,
		MaxRewardCashback:    maxReward,
		MaxRewardAllowedRate: rate,
		Rewards:              rewards,
	}, nil
}

// parseRat разбирает десятичную строку ("0.85") в точную дробь без float64
func parseRat(field, s string) (*big.Rat, error) {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, &apperr.ErrInvalidCampaign{Field: field, Reason: "must be a number, got " + strconv.Quote(s)}
	}

	return r, nil
}
