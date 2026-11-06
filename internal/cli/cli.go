// Package cli - консольный интерфейс: разбор аргументов и вывод через io.Writer
package cli

import (
	"bytes"
	"cashbackmerge/internal/config"
	"cashbackmerge/internal/features/bonus"
	"cashbackmerge/internal/features/featerr"
	"cashbackmerge/internal/features/game"
	"cashbackmerge/internal/features/simulation"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const usage = `Usage:
  cashback-merge new --config <path> [--seed <n>]
  cashback-merge simulate --config <path> [--players <n>] [--strategy random|greedy-score] [--seed <n>] [--output <path>]
  cashback-merge play --config <path> [--seed <n>] [--save <path>]
  cashback-merge replay --config <path> --input <path>
  cashback-merge compare --config <path> [--config-b <path>] [--strategy <name>] [--strategy-b <name>]
                         [--players <n>] [--seed <n>] [--output <path>]

Commands:
  new        загрузить кампанию и показать начальное поле по seed
  simulate   сыграть партии за players игроков и сохранить Markdown-отчёт в output (без output - в stdout)
  play       сыграть партию в консоли: ходы w/a/s/d или up/down/left/right, q - выход;
             с save сохранить seed и ходы партии в JSON
  replay     воспроизвести партию из файла play --save со всеми промежуточными состояниями
  compare    сравнить два варианта на одинаковых seed: две кампании (config и config-b)
             и/или две стратегии (strategy и strategy-b); без config-b и strategy-b берутся config и strategy
`

// reportFileMode - права на файлы отчёта и записи партии
const reportFileMode = 0o600

// usageError - неверные аргументы командной строки, к ним Run добавляет подсказку usage
type usageError struct {
	msg string
}

func (e *usageError) Error() string {
	return e.msg
}

// Run выполняет команду и возвращает код завершения: 0 - успех, 1 - ошибка.
// stdin нужен только консольной игре play
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	err := run(args, stdin, stdout)
	if err == nil {
		return 0
	}

	msg := "error: " + err.Error() + "\n"

	var usageErr *usageError
	if errors.As(err, &usageErr) {
		msg += "\n" + usage
	}

	// если не удалось записать даже в stderr, сообщить об ошибке уже некуда
	_, _ = io.WriteString(stderr, msg)

	return 1
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return &usageError{msg: "no command"}
	}

	switch args[0] {
	case "new":
		return runNew(args[1:], stdout)
	case "simulate":
		return runSimulate(args[1:], stdout)
	case "play":
		return runPlay(args[1:], stdin, stdout)
	case "replay":
		return runReplay(args[1:], stdout)
	case "compare":
		return runCompare(args[1:], stdout)
	case "help", "-h", "--help":
		return write(stdout, usage)
	default:
		return &usageError{msg: fmt.Sprintf("unknown command %q", args[0])}
	}
}

func runNew(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	configPath := fs.String("config", "", "путь к JSON-файлу кампании")
	seed := fs.Uint64("seed", 0, "seed для начального поля")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return write(stdout, usage)
		}

		return &usageError{msg: err.Error()}
	}

	if *configPath == "" {
		return &usageError{msg: "--config is required"}
	}

	campaign, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("load campaign: %w", err)
	}

	state, err := game.NewGame(campaign, game.NewRandom(*seed))
	if err != nil {
		return fmt.Errorf("new game: %w", err)
	}

	return write(stdout, fmt.Sprintf("Moves: %d/%d  Score: %d\n%s",
		state.MovesUsed, state.MovesLimit, state.Score, state.Board))
}

func runSimulate(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("simulate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	configPath := fs.String("config", "", "путь к JSON-файлу кампании")
	players := fs.Int("players", 1000, "число игроков")
	strategyName := fs.String("strategy", "random", "стратегия выбора хода: random или greedy-score")
	seed := fs.Uint64("seed", 0, "общий seed симуляции")
	output := fs.String("output", "", "путь к файлу Markdown-отчёта")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return write(stdout, usage)
		}

		return &usageError{msg: err.Error()}
	}

	if *configPath == "" {
		return &usageError{msg: "--config is required"}
	}

	if *players <= 0 {
		return &usageError{msg: "--players must be positive"}
	}

	strategy, err := strategyByName(*strategyName)
	if err != nil {
		return &usageError{msg: err.Error()}
	}

	campaign, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("load campaign: %w", err)
	}

	res, err := simulation.Run(campaign, simulation.Params{Players: *players, Seed: *seed, Strategy: strategy})
	if err != nil {
		return fmt.Errorf("simulate: %w", err)
	}

	if *output == "" {
		return simulation.WriteMarkdown(stdout, res)
	}

	var buf bytes.Buffer
	if err := simulation.WriteMarkdown(&buf, res); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Clean(*output), buf.Bytes(), reportFileMode); err != nil {
		return fmt.Errorf("save report: %w", err)
	}

	return write(stdout, fmt.Sprintf("Players: %d  Max reward rate: %s / %s  Status: %s\nReport: %s\n",
		res.Players, res.MaxRewardRate().FloatString(4), campaign.MaxRewardAllowedRate.FloatString(4),
		res.Status(), *output))
}

func runPlay(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("play", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	configPath := fs.String("config", "", "путь к JSON-файлу кампании")
	seed := fs.Uint64("seed", 0, "seed партии")
	save := fs.String("save", "", "путь к JSON-файлу записи партии для replay")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return write(stdout, usage)
		}

		return &usageError{msg: err.Error()}
	}

	if *configPath == "" {
		return &usageError{msg: "--config is required"}
	}

	campaign, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("load campaign: %w", err)
	}

	_, moves, err := bonus.Play(campaign, *seed, stdin, stdout)
	if err != nil {
		return fmt.Errorf("play: %w", err)
	}

	if *save == "" {
		return nil
	}

	var buf bytes.Buffer
	if err := bonus.WriteRecord(&buf, bonus.Record{Seed: *seed, Moves: moves}); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Clean(*save), buf.Bytes(), reportFileMode); err != nil {
		return fmt.Errorf("save record: %w", err)
	}

	return write(stdout, "Record: "+*save+"\n")
}

func runReplay(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	configPath := fs.String("config", "", "путь к JSON-файлу кампании")
	input := fs.String("input", "", "путь к JSON-файлу записи партии")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return write(stdout, usage)
		}

		return &usageError{msg: err.Error()}
	}

	if *configPath == "" {
		return &usageError{msg: "--config is required"}
	}

	if *input == "" {
		return &usageError{msg: "--input is required"}
	}

	campaign, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("load campaign: %w", err)
	}

	data, err := os.ReadFile(filepath.Clean(*input))
	if err != nil {
		return fmt.Errorf("read record: %w", err)
	}

	rec, err := bonus.ReadRecord(bytes.NewReader(data))
	if err != nil {
		return err
	}

	states, err := bonus.Replay(campaign, rec)
	if err != nil {
		return fmt.Errorf("replay: %w", err)
	}

	return bonus.WriteReplay(stdout, campaign, rec, states)
}

func runCompare(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	configA := fs.String("config", "", "путь к JSON-файлу кампании A")
	configB := fs.String("config-b", "", "путь к JSON-файлу кампании B, по умолчанию config")
	strategyA := fs.String("strategy", "random", "стратегия варианта A: random или greedy-score")
	strategyB := fs.String("strategy-b", "", "стратегия варианта B, по умолчанию strategy")
	players := fs.Int("players", 1000, "число игроков в каждом варианте")
	seed := fs.Uint64("seed", 0, "общий seed обоих вариантов")
	output := fs.String("output", "", "путь к файлу Markdown-отчёта")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return write(stdout, usage)
		}

		return &usageError{msg: err.Error()}
	}

	if *configA == "" {
		return &usageError{msg: "--config is required"}
	}

	if *players <= 0 {
		return &usageError{msg: "--players must be positive"}
	}

	if *configB == "" {
		*configB = *configA
	}

	if *strategyB == "" {
		*strategyB = *strategyA
	}

	var variants [2]bonus.Variant

	for i, v := range []struct{ config, strategy string }{{*configA, *strategyA}, {*configB, *strategyB}} {
		strategy, err := strategyByName(v.strategy)
		if err != nil {
			return &usageError{msg: err.Error()}
		}

		campaign, err := config.Load(v.config)
		if err != nil {
			return fmt.Errorf("load campaign %s: %w", v.config, err)
		}

		variants[i] = bonus.Variant{Name: v.config + ", " + v.strategy, Campaign: campaign, Strategy: strategy}
	}

	cmp, err := bonus.Compare(variants[0], variants[1], *players, *seed)
	if err != nil {
		return fmt.Errorf("compare: %w", err)
	}

	if *output == "" {
		return bonus.WriteCompareMarkdown(stdout, cmp)
	}

	var buf bytes.Buffer
	if err := bonus.WriteCompareMarkdown(&buf, cmp); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Clean(*output), buf.Bytes(), reportFileMode); err != nil {
		return fmt.Errorf("save report: %w", err)
	}

	return write(stdout, fmt.Sprintf("A: %s  Status: %s\nB: %s  Status: %s\nReport: %s\n",
		cmp.Names[0], cmp.Results[0].Status(), cmp.Names[1], cmp.Results[1].Status(), *output))
}

// strategyByName возвращает стратегию по флагу --strategy
func strategyByName(name string) (game.Strategy, error) {
	switch name {
	case game.RandomStrategy{}.Name():
		return game.RandomStrategy{}, nil
	case bonus.ScoreStrategy{}.Name():
		return bonus.ScoreStrategy{}, nil
	default:
		return nil, fmt.Errorf("%w %q", featerr.ErrUnknownStrategy, name)
	}
}

func write(w io.Writer, s string) error {
	if _, err := io.WriteString(w, s); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
