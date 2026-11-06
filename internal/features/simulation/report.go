// Package simulation - batch-симуляция кампании: прогон партий, сбор статистики и отчёт
package simulation

import (
	"cashbackmerge/internal/core/domain"
	"fmt"
	"io"
	"math/big"
	"strings"
)

// WriteMarkdown пишет отчёт симуляции в Markdown
func WriteMarkdown(w io.Writer, r Result) error {
	var sb strings.Builder

	c := r.Campaign

	sb.WriteString("# Отчёт симуляции Cashback Merge\n\n")

	sb.WriteString("## Запуск\n\n")
	sb.WriteString("| Параметр | Значение |\n| --- | --- |\n")
	fmt.Fprintf(&sb, "| Игроков | %d |\n", r.Players)
	fmt.Fprintf(&sb, "| Лимит ходов | %d |\n", c.MovesLimit)
	fmt.Fprintf(&sb, "| Seed | %d |\n", r.Seed)
	fmt.Fprintf(&sb, "| Стратегия | %s |\n\n", r.Strategy)

	sb.WriteString("## Кампания\n\n")
	sb.WriteString("| Новая плитка | Вероятность |\n| --- | ---: |\n")

	for _, rule := range c.SpawnRules {
		fmt.Fprintf(&sb, "| %s | %s |\n", rule.Cashback, percent(rule.Probability))
	}

	fmt.Fprintf(&sb, "\nПлитка максимальной награды: %s\n\n", c.MaxRewardCashback)

	sb.WriteString("## Распределение максимальных плиток\n\n")
	sb.WriteString("| Максимальная плитка | Игроков | Доля |\n| --- | ---: | ---: |\n")

	for _, tile := range sortedTiles(r.TileCounts) {
		n := r.TileCounts[tile]
		fmt.Fprintf(&sb, "| %s | %d | %s |\n", tile, n, share(n, r.Players))
	}

	sb.WriteString("\n## Распределение наград\n\n")
	sb.WriteString("| Награда | Минимальная плитка | Игроков | Доля |\n| --- | --- | ---: | ---: |\n")

	// уровни выводятся в порядке таблицы наград, включая уровни без игроков
	for _, reward := range c.Rewards {
		n := r.RewardCounts[reward.Name]
		fmt.Fprintf(&sb, "| %s | %s | %d | %s |\n", reward.Name, reward.MinTile, n, share(n, r.Players))
	}

	sb.WriteString("\n## Score\n\n")
	sb.WriteString("| Показатель | Значение |\n| --- | ---: |\n")
	fmt.Fprintf(&sb, "| Средний score | %s |\n", r.AverageScore().FloatString(2))
	fmt.Fprintf(&sb, "| Максимальный score | %d |\n\n", r.MaxScore)

	sb.WriteString("## Ограничение максимальной награды\n\n")
	sb.WriteString("| Показатель | Значение |\n| --- | ---: |\n")
	fmt.Fprintf(&sb, "| Игроков с максимальной наградой | %d |\n", r.MaxRewardPlayers)
	fmt.Fprintf(&sb, "| Фактическая доля | %s |\n", percent(r.MaxRewardRate()))
	fmt.Fprintf(&sb, "| Допустимая доля | %s |\n", percent(c.MaxRewardAllowedRate))
	fmt.Fprintf(&sb, "| Статус | **%s** |\n\n", r.Status())

	if _, err := io.WriteString(w, sb.String()); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	return nil
}

// sortedTiles возвращает плитки из counts по возрастанию
func sortedTiles(counts map[domain.Cashback]int) []domain.Cashback {
	tiles := make([]domain.Cashback, 0, len(counts))

	for p := uint(1); p != 0 && len(tiles) < len(counts); p <<= 1 {
		tile, err := domain.NewCashback(p)
		if err != nil {
			continue
		}

		if _, ok := counts[tile]; ok {
			tiles = append(tiles, tile)
		}
	}

	return tiles
}

// share - доля n от total в процентах
func share(n, total int) string {
	return percent(big.NewRat(int64(n), int64(total)))
}

// percent выводит дробь в процентах с двумя знаками без float
func percent(r *big.Rat) string {
	return new(big.Rat).Mul(r, big.NewRat(100, 1)).FloatString(2) + "%"
}
