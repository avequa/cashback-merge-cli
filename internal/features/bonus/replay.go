package bonus

import (
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/featerr"
	"cashbackmerge/internal/features/game"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Record - запись партии для воспроизведения: seed и ходы, которые изменили поле, в порядке игры
type Record struct {
	Seed  uint64
	Moves []domain.Direction
}

// recordJSON - формат файла записи: {"seed": 42, "moves": ["UP", "LEFT"]}
type recordJSON struct {
	Seed  uint64   `json:"seed"`
	Moves []string `json:"moves"`
}

// recordNames - названия направлений в файле записи
var recordNames = map[domain.Direction]string{
	domain.Up:    "UP",
	domain.Down:  "DOWN",
	domain.Left:  "LEFT",
	domain.Right: "RIGHT",
}

// WriteRecord сохраняет запись партии в JSON
func WriteRecord(w io.Writer, rec Record) error {
	raw := recordJSON{Seed: rec.Seed, Moves: make([]string, 0, len(rec.Moves))}

	for _, d := range rec.Moves {
		name, ok := recordNames[d]
		if !ok {
			return fmt.Errorf("record move: invalid direction %d", d)
		}

		raw.Moves = append(raw.Moves, name)
	}

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("encode record: %w", err)
	}

	return write(w, string(data)+"\n")
}

// ReadRecord читает запись партии из JSON, неизвестные поля и направления считаются ошибкой
func ReadRecord(r io.Reader) (Record, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var raw recordJSON
	if err := dec.Decode(&raw); err != nil {
		return Record{}, fmt.Errorf("decode record: %w", err)
	}

	rec := Record{Seed: raw.Seed, Moves: make([]domain.Direction, 0, len(raw.Moves))}

	for i, name := range raw.Moves {
		d, ok := ParseDirection(name)
		if !ok {
			return Record{}, fmt.Errorf("decode record: moves[%d]: unknown direction %q", i, name)
		}

		rec.Moves = append(rec.Moves, d)
	}

	return rec, nil
}

// Replay воспроизводит партию по кампании и записи. Возвращает все состояния: начальное и после каждого хода, последнее - итог партии
func Replay(c domain.Campaign, rec Record) ([]domain.State, error) {
	rng := game.NewRandom(rec.Seed)

	state, err := game.NewGame(c, rng)
	if err != nil {
		return nil, err
	}

	states := make([]domain.State, 0, len(rec.Moves)+1)
	states = append(states, state)

	for i, d := range rec.Moves {
		res, err := game.ApplyMove(state, d, c.SpawnRules, rng)
		if err != nil {
			return nil, fmt.Errorf("move %d: %w", i+1, err)
		}

		if !res.Changed {
			return nil, fmt.Errorf("move %d %s: %w", i+1, recordNames[d], featerr.ErrRecordMismatch)
		}

		state = res.State
		states = append(states, state)
	}

	return states, nil
}

// WriteReplay выводит воспроизведённую партию: каждое состояние в формате консольной игры
func WriteReplay(w io.Writer, c domain.Campaign, rec Record, states []domain.State) error {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Replay: seed %d, %d moves\n", rec.Seed, len(rec.Moves))
	sb.WriteString(status(c, states[0]))

	for i, s := range states[1:] {
		fmt.Fprintf(&sb, "\n> %s", recordNames[rec.Moves[i]])
		sb.WriteString(status(c, s))
	}

	if last := states[len(states)-1]; last.Finished {
		sb.WriteString("Game over\n")
	}

	return write(w, sb.String())
}
