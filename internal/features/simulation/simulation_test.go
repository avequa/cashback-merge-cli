package simulation

import (
	"bytes"
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/game"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func cashback(t *testing.T, s string) domain.Cashback {
	t.Helper()

	c, err := domain.ParseCashback(s)
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func testCampaign(t *testing.T) domain.Campaign {
	t.Helper()

	return domain.Campaign{
		BoardSize:  4,
		MovesLimit: 100,
		SpawnRules: []domain.SpawnRule{
			{Cashback: cashback(t, "1"), Probability: big.NewRat(85, 100)},
			{Cashback: cashback(t, "2"), Probability: big.NewRat(15, 100)},
		},
		MaxRewardCashback:    cashback(t, "32"),
		MaxRewardAllowedRate: big.NewRat(20, 100),
		Rewards: []domain.Reward{
			{MinTile: cashback(t, "1"), Name: "basic"},
			{MinTile: cashback(t, "8"), Name: "silver"},
			{MinTile: cashback(t, "16"), Name: "gold"},
			{MinTile: cashback(t, "32"), Name: "platinum"},
		},
	}
}

func TestRunSameSeedSameResult(t *testing.T) {
	p := Params{Players: 200, Seed: 42, Strategy: game.RandomStrategy{}}

	first, err := Run(testCampaign(t), p)
	if err != nil {
		t.Fatal(err)
	}

	second, err := Run(testCampaign(t), p)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed gave different results:\n%+v\n%+v", first, second)
	}
}

func TestRunCountsEveryPlayer(t *testing.T) {
	res, err := Run(testCampaign(t), Params{Players: 200, Seed: 42, Strategy: game.RandomStrategy{}})
	if err != nil {
		t.Fatal(err)
	}

	sum := func(m map[string]int) int {
		n := 0
		for _, v := range m {
			n += v
		}

		return n
	}

	tiles := 0
	for _, v := range res.TileCounts {
		tiles += v
	}

	if tiles != 200 || sum(res.RewardCounts) != 200 {
		t.Errorf("tiles = %d, rewards = %d, want 200 and 200", tiles, sum(res.RewardCounts))
	}

	if res.MaxScore <= 0 || res.TotalScore < res.MaxScore {
		t.Errorf("MaxScore = %d, TotalScore = %d", res.MaxScore, res.TotalScore)
	}

	if res.Strategy != "random" || res.Seed != 42 || res.Players != 200 {
		t.Errorf("run params not saved: %+v", res)
	}
}

// TestStatus проверяет статус по обе стороны лимита 0.20 на 10 игроках
func TestStatus(t *testing.T) {
	tests := []struct {
		name             string
		maxRewardPlayers int
		want             Status
	}{
		{"below limit", 1, StatusOK},
		{"exactly at limit", 2, StatusOK},
		{"above limit", 3, StatusFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Result{Campaign: testCampaign(t), Players: 10, MaxRewardPlayers: tt.maxRewardPlayers}

			if got := res.Status(); got != tt.want {
				t.Errorf("Status() = %s, want %s (rate %s)", got, tt.want, res.MaxRewardRate().RatString())
			}
		})
	}
}

func render(t *testing.T, r Result) string {
	t.Helper()

	var buf bytes.Buffer
	if err := WriteMarkdown(&buf, r); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}

func runReport(t *testing.T, seed uint64) string {
	t.Helper()

	res, err := Run(testCampaign(t), Params{
		Players:  100,
		Seed:     seed,
		Strategy: game.RandomStrategy{},
	})
	if err != nil {
		t.Fatal(err)
	}

	return render(t, res)
}

func TestWriteMarkdownSameSeedSameReport(t *testing.T) {
	if first, second := runReport(t, 42), runReport(t, 42); first != second {
		t.Fatalf("same seed gave different reports:\n%s\n---\n%s", first, second)
	}
}

func TestWriteMarkdownContainsRequiredFields(t *testing.T) {
	c := testCampaign(t)

	res := Result{
		Campaign: c,
		Players:  10,
		Seed:     42,
		Strategy: "random",
		TileCounts: map[domain.Cashback]int{
			cashback(t, "32"): 3,
			cashback(t, "8"):  4,
			cashback(t, "16"): 3,
		},
		RewardCounts:     map[string]int{"basic": 4, "gold": 3, "platinum": 3},
		TotalScore:       1234,
		MaxScore:         300,
		MaxRewardPlayers: 3,
	}

	got := render(t, res)

	want := []string{
		"| Игроков | 10 |",
		"| Лимит ходов | 100 |",
		"| Seed | 42 |",
		"| Стратегия | random |",
		"| 1% | 85.00% |",
		"## Распределение максимальных плиток",
		"| 16% | 3 | 30.00% |",
		"## Распределение наград",
		"| gold | 16% | 3 | 30.00% |",
		"| Средний score | 123.40 |",
		"| Максимальный score | 300 |",
		"| Игроков с максимальной наградой | 3 |",
		"| Фактическая доля | 30.00% |",
		"| Допустимая доля | 20.00% |",
		"| Статус | **FAILED** |",
	}

	for _, s := range want {
		if !strings.Contains(got, s) {
			t.Errorf("report does not contain %q:\n%s", s, got)
		}
	}

	// плитки выводятся по возрастанию независимо от порядка обхода map
	if i8, i16, i32 := strings.Index(got, "| 8% |"), strings.Index(got, "| 16% |"), strings.Index(got, "| 32% |"); i8 >= i16 || i16 >= i32 {
		t.Errorf("tiles are not sorted:\n%s", got)
	}

	res.MaxRewardPlayers = 2
	if got := render(t, res); !strings.Contains(got, "| Статус | **OK** |") {
		t.Errorf("report does not contain OK status:\n%s", got)
	}
}
