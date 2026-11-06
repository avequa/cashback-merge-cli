package config

import (
	"cashbackmerge/internal/core/apperr"
	"errors"
	"strings"
	"testing"
)

const validConfig = `{
  "movesLimit": 100,
  "spawnRules": [{"cashback": "1", "probability": "0.85"}, {"cashback": "2", "probability": "0.15"}],
  "maxRewardCashback": "32",
  "maxRewardAllowedRate": "0.20",
  "rewards": [
    {"minTile": "1", "name": "basic"},
    {"minTile": "16", "name": "gold"},
    {"minTile": "32", "name": "platinum"}
  ]
}`

func TestParseValidConfig(t *testing.T) {
	c, err := Parse(strings.NewReader(validConfig))
	if err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	if c.MovesLimit != 100 {
		t.Errorf("MovesLimit = %d, want 100", c.MovesLimit)
	}
}

func TestParseInvalidConfig(t *testing.T) {
	tests := []struct {
		name      string
		old, new  string
		wantField string
	}{
		{
			name:      "empty spawn rules",
			old:       `[{"cashback": "1", "probability": "0.85"}, {"cashback": "2", "probability": "0.15"}]`,
			new:       `[]`,
			wantField: "spawnRules",
		},
		{
			name:      "probabilities do not sum to 1",
			old:       `"0.15"`,
			new:       `"0.10"`,
			wantField: "spawnRules",
		},
		{
			name:      "zero moves limit",
			old:       `"movesLimit": 100`,
			new:       `"movesLimit": 0`,
			wantField: "movesLimit",
		},
		{
			name:      "negative moves limit",
			old:       `"movesLimit": 100`,
			new:       `"movesLimit": -1`,
			wantField: "movesLimit",
		},
		{
			name:      "negative probability",
			old:       `"0.85"`,
			new:       `"-0.85"`,
			wantField: "spawnRules[0].probability",
		},
		{
			name:      "probability above 1",
			old:       `"0.15"`,
			new:       `"1.15"`,
			wantField: "spawnRules[1].probability",
		},
		{
			name:      "rewards lowest level above smallest spawn",
			old:       `{"minTile": "1", "name": "basic"}`,
			new:       `{"minTile": "2", "name": "basic"}`,
			wantField: "rewards",
		},
		{
			name:      "rewards highest level differs from maxRewardCashback",
			old:       `"maxRewardCashback": "32"`,
			new:       `"maxRewardCashback": "64"`,
			wantField: "rewards",
		},
		{
			name:      "rewards minTile not increasing",
			old:       `"minTile": "16"`,
			new:       `"minTile": "1"`,
			wantField: "rewards[1].minTile",
		},
		{
			name:      "rewards empty name",
			old:       `"name": "gold"`,
			new:       `"name": ""`,
			wantField: "rewards[1].name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(strings.Replace(validConfig, tt.old, tt.new, 1)))

			var campaignErr *apperr.ErrInvalidCampaign
			if !errors.As(err, &campaignErr) {
				t.Fatalf("err = %v, want *apperr.ErrInvalidCampaign", err)
			}

			if campaignErr.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", campaignErr.Field, tt.wantField)
			}
		})
	}
}

func TestParseInvalidCashback(t *testing.T) {
	tests := []struct {
		name     string
		old, new string
	}{
		{"not power of two", `"cashback": "1"`, `"cashback": "3"`},
		{"zero", `"cashback": "1"`, `"cashback": "0"`},
		{"not a number", `"cashback": "1"`, `"cashback": "abc"`},
		{"invalid maxRewardCashback", `"maxRewardCashback": "32"`, `"maxRewardCashback": "33"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(strings.Replace(validConfig, tt.old, tt.new, 1)))

			var cashbackErr *apperr.ErrInvalidCashback[string]
			if !errors.As(err, &cashbackErr) {
				t.Fatalf("err = %v, want *apperr.ErrInvalidCashback[string]", err)
			}
		})
	}
}

func TestParseRejectsTrailingData(t *testing.T) {
	if _, err := Parse(strings.NewReader(validConfig + ` {}`)); err == nil {
		t.Fatal("config with trailing data accepted")
	}
}
