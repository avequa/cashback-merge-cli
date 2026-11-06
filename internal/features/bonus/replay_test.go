package bonus

import (
	"bytes"
	"cashbackmerge/internal/core/domain"
	"cashbackmerge/internal/features/featerr"
	"cashbackmerge/internal/features/game"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestReplayRepeatsPlayedGame(t *testing.T) {
	c, err := parseCampaign(t, "")
	if err != nil {
		t.Fatal(err)
	}

	c.MovesLimit = 10

	var out strings.Builder

	final, moves, err := Play(c, 42, strings.NewReader(strings.Repeat("w\na\ns\nd\n", 10)), &out)
	if err != nil {
		t.Fatal(err)
	}

	states, err := Replay(c, Record{Seed: 42, Moves: moves})
	if err != nil {
		t.Fatal(err)
	}

	if len(states) != len(moves)+1 {
		t.Fatalf("got %d states, want %d", len(states), len(moves)+1)
	}

	if last := states[len(states)-1]; last != final || !last.Finished {
		t.Fatalf("replay result differs from played game:\n%s\n%s", last.Board, final.Board)
	}

	// каждое промежуточное состояние replay совпадает с тем, что игрок видел в консоли, и в том же порядке
	played := out.String()

	for i, s := range states {
		pos := strings.Index(played, status(c, s))
		if pos < 0 {
			t.Fatalf("state %d of replay was not shown in played game:\n%s", i, status(c, s))
		}

		played = played[pos+1:]
	}
}

func TestRecordJSONRoundTrip(t *testing.T) {
	rec := Record{Seed: 42, Moves: []domain.Direction{domain.Up, domain.Down, domain.Left, domain.Right}}

	var buf bytes.Buffer
	if err := WriteRecord(&buf, rec); err != nil {
		t.Fatal(err)
	}

	got, err := ReadRecord(&buf)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, rec) {
		t.Fatalf("got %+v, want %+v", got, rec)
	}
}

func TestReadRecordInvalid(t *testing.T) {
	for _, data := range []string{
		`{"seed": 42, "moves": ["UP", "JUMP"]}`,
		`{"seed": 42, "moves": ["UP"], "score": 10}`,
		`{"seed": -1, "moves": []}`,
	} {
		if _, err := ReadRecord(strings.NewReader(data)); err == nil {
			t.Errorf("ReadRecord(%s): got nil error", data)
		}
	}
}

func TestReplayRejectsMoveWithoutChange(t *testing.T) {
	c, err := parseCampaign(t, "")
	if err != nil {
		t.Fatal(err)
	}

	start, err := game.NewGame(c, game.NewRandom(42))
	if err != nil {
		t.Fatal(err)
	}

	for _, d := range domain.Directions() {
		next, _, err := start.Board.MergeBoardAfterMove(d)
		if err != nil {
			t.Fatal(err)
		}

		if next != start.Board {
			continue
		}

		if _, err := Replay(c, Record{Seed: 42, Moves: []domain.Direction{d}}); !errors.Is(err, featerr.ErrRecordMismatch) {
			t.Fatalf("got error %v, want ErrRecordMismatch", err)
		}

		return
	}

	t.Fatalf("every direction changes the start board:\n%s", start.Board)
}

func TestWriteReplay(t *testing.T) {
	c, err := parseCampaign(t, "")
	if err != nil {
		t.Fatal(err)
	}

	c.MovesLimit = 2

	_, moves, err := Play(c, 7, strings.NewReader(strings.Repeat("w\na\ns\nd\n", 2)), &strings.Builder{})
	if err != nil {
		t.Fatal(err)
	}

	rec := Record{Seed: 7, Moves: moves}

	states, err := Replay(c, rec)
	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	if err := WriteReplay(&out, c, rec, states); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"Replay: seed 7, 2 moves", "> " + recordNames[moves[0]], "Moves: 2/2  Left: 0", "Game over"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output does not contain %q:\n%s", want, out.String())
		}
	}
}
