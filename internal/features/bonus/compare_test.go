package bonus

import (
	"cashbackmerge/internal/features/game"
	"reflect"
	"strings"
	"testing"
)

func compareVariants(t *testing.T) (Variant, Variant) {
	t.Helper()

	a, err := parseCampaign(t, "")
	if err != nil {
		t.Fatal(err)
	}

	b, err := parseCampaign(t, `"boardSize": 5,`)
	if err != nil {
		t.Fatal(err)
	}

	b.MovesLimit = 20

	return Variant{Name: "a.json, random", Campaign: a, Strategy: game.RandomStrategy{}},
		Variant{Name: "b.json, greedy-score", Campaign: b, Strategy: ScoreStrategy{}}
}

func TestCompareSameSeedSameResult(t *testing.T) {
	a, b := compareVariants(t)

	first, err := Compare(a, b, 100, 42)
	if err != nil {
		t.Fatal(err)
	}

	second, err := Compare(a, b, 100, 42)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed gave different comparisons")
	}
}

func TestCompareSameVariantGivesEqualResults(t *testing.T) {
	a, _ := compareVariants(t)

	cmp, err := Compare(a, a, 100, 42)
	if err != nil {
		t.Fatal(err)
	}

	// одинаковые seed: один и тот же вариант дважды даёт одинаковую статистику
	if !reflect.DeepEqual(cmp.Results[0], cmp.Results[1]) {
		t.Fatalf("same variant gave different results:\n%+v\n%+v", cmp.Results[0], cmp.Results[1])
	}
}

func TestWriteCompareMarkdown(t *testing.T) {
	a, b := compareVariants(t)

	cmp, err := Compare(a, b, 100, 42)
	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	if err := WriteCompareMarkdown(&out, cmp); err != nil {
		t.Fatal(err)
	}

	got := out.String()

	want := []string{
		"| Игроков в каждом варианте | 100 |",
		"| Seed | 42 |",
		"| Показатель | A: a.json, random | B: b.json, greedy-score |",
		"| Стратегия | random | greedy-score |",
		"| Размер поля | 4×4 | 5×5 |",
		"| Лимит ходов | 100 | 20 |",
		"| Средний score | " + cmp.Results[0].AverageScore().FloatString(2) + " | " + cmp.Results[1].AverageScore().FloatString(2) + " |",
		"| Доля максимальной награды | " + percent(cmp.Results[0].MaxRewardRate()) + " | " + percent(cmp.Results[1].MaxRewardRate()) + " |",
		"| Статус | **" + string(cmp.Results[0].Status()) + "** | **" + string(cmp.Results[1].Status()) + "** |",
		"## Распределение наград",
		"### A: a.json, random",
		"### B: b.json, greedy-score",
		"| platinum | 32% |",
	}

	for _, s := range want {
		if !strings.Contains(got, s) {
			t.Errorf("report does not contain %q:\n%s", s, got)
		}
	}
}
