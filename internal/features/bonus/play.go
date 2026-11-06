// Package bonus - реализация бонусной части
package bonus

import (
	"bufio"
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/game"
	"fmt"
	"io"
	"strings"
)

// playHelp - подсказка по командам консольной игры
const playHelp = "Moves: w/up, s/down, a/left, d/right; q - quit\n"

// directionNames - команды консольной игры для каждого направления
var directionNames = map[string]domain.Direction{
	"w": domain.Up, "up": domain.Up,
	"s": domain.Down, "down": domain.Down,
	"a": domain.Left, "left": domain.Left,
	"d": domain.Right, "right": domain.Right,
}

// ParseDirection разбирает команду хода без учёта регистра: w/up, s/down, a/left, d/right
func ParseDirection(s string) (domain.Direction, bool) {
	d, ok := directionNames[strings.ToLower(strings.TrimSpace(s))]

	return d, ok
}

// Play - консольная партия: читает ходы построчно из in, после каждого хода пишет в out
// поле, использованные и оставшиеся ходы, score и текущую награду.
// Партия заканчивается, когда она завершена по правилам или по команде q
// Возвращает итоговое состояние и ходы, которые изменили поле, в порядке ввода
func Play(c domain.Campaign, seed uint64, in io.Reader, out io.Writer) (domain.State, []domain.Direction, error) {
	rng := game.NewRandom(seed)

	state, err := game.NewGame(c, rng)
	if err != nil {
		return domain.State{}, nil, err
	}

	var moves []domain.Direction

	if err := write(out, playHelp+status(c, state)); err != nil {
		return domain.State{}, nil, err
	}

	scanner := bufio.NewScanner(in)

	for !state.Finished && scanner.Scan() {
		cmd := strings.TrimSpace(scanner.Text())
		if strings.EqualFold(cmd, "q") {
			break
		}

		d, ok := ParseDirection(cmd)
		if !ok {
			if err := write(out, fmt.Sprintf("unknown command %q\n%s", cmd, playHelp)); err != nil {
				return domain.State{}, nil, err
			}

			continue
		}

		res, err := game.ApplyMove(state, d, c.SpawnRules, rng)
		if err != nil {
			return domain.State{}, nil, fmt.Errorf("apply move: %w", err)
		}

		if !res.Changed {
			if err := write(out, "move does not change the board, try another direction\n"); err != nil {
				return domain.State{}, nil, err
			}

			continue
		}

		state = res.State
		moves = append(moves, d)

		if err := write(out, status(c, state)); err != nil {
			return domain.State{}, nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.State{}, nil, fmt.Errorf("read move: %w", err)
	}

	if state.Finished {
		if err := write(out, "Game over\n"); err != nil {
			return domain.State{}, nil, err
		}
	}

	return state, moves, nil
}

// status - поле, ходы, score и текущая награда по максимальной плитке
func status(c domain.Campaign, s domain.State) string {
	return fmt.Sprintf("\n%sMoves: %d/%d  Left: %d  Score: %d  Reward: %s\n",
		s.Board, s.MovesUsed, s.MovesLimit, s.MovesLimit-s.MovesUsed, s.Score,
		c.RewardFor(s.Board.MaxTile()).Name)
}

func write(w io.Writer, s string) error {
	if _, err := io.WriteString(w, s); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
