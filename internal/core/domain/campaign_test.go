package domain

import "testing"

func rewardCampaign() Campaign {
	return Campaign{
		MaxRewardCashback: Cashback{percent: 32},
		Rewards: []Reward{
			{MinTile: Cashback{percent: 1}, Name: "basic"},
			{MinTile: Cashback{percent: 4}, Name: "bronze"},
			{MinTile: Cashback{percent: 8}, Name: "silver"},
			{MinTile: Cashback{percent: 16}, Name: "gold"},
			{MinTile: Cashback{percent: 32}, Name: "platinum"},
		},
	}
}

func TestRewardFor(t *testing.T) {
	tests := []struct {
		name    string
		maxTile uint
		want    string
		wantMax bool
	}{
		{"lowest level", 1, "basic", false},
		{"between levels", 2, "basic", false},
		{"exact level", 4, "bronze", false},
		{"16 gives gold", 16, "gold", false},
		{"32 gives max reward", 32, "platinum", true},
		{"above max reward", 64, "platinum", true},
	}

	c := rewardCampaign()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := Cashback{percent: tt.maxTile}

			if got := c.RewardFor(tile).Name; got != tt.want {
				t.Errorf("RewardFor(%s) = %q, want %q", tile, got, tt.want)
			}

			if got := c.IsMaxReward(tile); got != tt.wantMax {
				t.Errorf("IsMaxReward(%s) = %t, want %t", tile, got, tt.wantMax)
			}
		})
	}
}

func TestMaxTile(t *testing.T) {
	b := boardOf([][]uint{
		{1, 0, 0, 2},
		{0, 16, 0, 0},
		{4, 0, 8, 0},
	})

	if got := b.MaxTile(); got != (Cashback{percent: 16}) {
		t.Errorf("MaxTile() = %s, want 16%%", got)
	}

	if got := (Board{}).MaxTile(); got != (Cashback{}) {
		t.Errorf("empty board MaxTile() = %s, want zero value", got)
	}
}
