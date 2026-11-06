package bonus

import (
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/game"
	"cashbackmerge/internal/features/simulation"
	"fmt"
	"io"
	"math/big"
	"strings"
)

// Variant - один из сравниваемых вариантов: кампания и стратегия
type Variant struct {
	Name     string // подпись в отчёте, например путь к конфигу
	Campaign domain.Campaign
	Strategy game.Strategy
}

// Comparison - результаты двух вариантов на одинаковых seed
type Comparison struct {
	Players int
	Seed    uint64
	Names   [2]string
	Results [2]simulation.Result
}

// Compare запускает симуляцию для двух вариантов с одним числом игроков и одним seed:
// у i-го игрока в обоих вариантах один и тот же seed партии, поэтому разница в результатах -
// только от кампании и стратегии
func Compare(a, b Variant, players int, seed uint64) (Comparison, error) {
	cmp := Comparison{Players: players, Seed: seed, Names: [2]string{a.Name, b.Name}}

	for i, v := range []Variant{a, b} {
		res, err := simulation.Run(v.Campaign, simulation.Params{Players: players, Seed: seed, Strategy: v.Strategy})
		if err != nil {
			return Comparison{}, fmt.Errorf("variant %s: %w", v.Name, err)
		}

		cmp.Results[i] = res
	}

	return cmp, nil
}

// WriteCompareMarkdown пишет общий отчёт сравнения в Markdown
func WriteCompareMarkdown(w io.Writer, cmp Comparison) error {
	var sb strings.Builder

	a, b := cmp.Results[0], cmp.Results[1]

	sb.WriteString("# Сравнение кампаний Cashback Merge\n\n")
	sb.WriteString("## Запуск\n\n")
	sb.WriteString("| Параметр | Значение |\n| --- | --- |\n")
	fmt.Fprintf(&sb, "| Игроков в каждом варианте | %d |\n", cmp.Players)
	fmt.Fprintf(&sb, "| Seed | %d |\n\n", cmp.Seed)

	sb.WriteString("## Итоги\n\n")
	fmt.Fprintf(&sb, "| Показатель | A: %s | B: %s |\n| --- | ---: | ---: |\n", cmp.Names[0], cmp.Names[1])

	rows := []struct {
		name string
		a, b string
	}{
		{"Стратегия", a.Strategy, b.Strategy},
		{"Размер поля", boardSize(a.Campaign), boardSize(b.Campaign)},
		{"Лимит ходов", fmt.Sprint(a.Campaign.MovesLimit), fmt.Sprint(b.Campaign.MovesLimit)},
		{"Средний score", a.AverageScore().FloatString(2), b.AverageScore().FloatString(2)},
		{"Максимальный score", fmt.Sprint(a.MaxScore), fmt.Sprint(b.MaxScore)},
		{"Плитка максимальной награды", a.Campaign.MaxRewardCashback.String(), b.Campaign.MaxRewardCashback.String()},
		{"Игроков с максимальной наградой", fmt.Sprint(a.MaxRewardPlayers), fmt.Sprint(b.MaxRewardPlayers)},
		{"Доля максимальной награды", percent(a.MaxRewardRate()), percent(b.MaxRewardRate())},
		{"Допустимая доля", percent(a.Campaign.MaxRewardAllowedRate), percent(b.Campaign.MaxRewardAllowedRate)},
		{"Статус", "**" + string(a.Status()) + "**", "**" + string(b.Status()) + "**"},
	}

	for _, r := range rows {
		fmt.Fprintf(&sb, "| %s | %s | %s |\n", r.name, r.a, r.b)
	}

	// таблицы наград у кампаний могут различаться, поэтому распределение выводится для каждого варианта
	sb.WriteString("\n## Распределение наград\n")

	for i, res := range cmp.Results {
		fmt.Fprintf(&sb, "\n### %s: %s\n\n", string(rune('A'+i)), cmp.Names[i])
		sb.WriteString("| Награда | Минимальная плитка | Игроков | Доля |\n| --- | --- | ---: | ---: |\n")

		for _, reward := range res.Campaign.Rewards {
			n := res.RewardCounts[reward.Name]
			fmt.Fprintf(&sb, "| %s | %s | %d | %s |\n",
				reward.Name, reward.MinTile, n, percent(big.NewRat(int64(n), int64(res.Players))))
		}
	}

	return write(w, sb.String())
}

func boardSize(c domain.Campaign) string {
	return fmt.Sprintf("%d×%d", c.BoardSize, c.BoardSize)
}

// percent выводит дробь в процентах с двумя знаками без float
func percent(r *big.Rat) string {
	return new(big.Rat).Mul(r, big.NewRat(100, 1)).FloatString(2) + "%"
}
