package game

import (
	"testing"

	"xiangqi-lab/internal/domain/xiangqi"
)

func TestSnapshotIncludesAuthoritativeCheckState(t *testing.T) {
	match := Match{
		FEN: "4k4/9/9/9/4R4/9/9/9/9/3K5 b",
		Moves: []MoveRecord{{
			Move:     "d5e5",
			FENAfter: "4k4/9/9/9/4R4/9/9/9/9/3K5 b",
		}},
		Outcome:     xiangqi.OutcomeOngoing,
		Termination: xiangqi.TerminationNone,
	}

	snapshot := match.Snapshot()
	if !snapshot.InCheck {
		t.Fatal("snapshot must report that the side to move is in check")
	}
	if len(snapshot.Moves) != 1 || !snapshot.Moves[0].GivesCheck {
		t.Fatal("snapshot must annotate checking moves from persisted FEN data")
	}
}

func TestAppendMoveRecordsWhetherItGivesCheck(t *testing.T) {
	before, err := xiangqi.ParseFEN("4k4/9/9/9/3R5/9/9/9/9/3K5 w")
	if err != nil {
		t.Fatal(err)
	}
	move, err := xiangqi.ParseMove("d5e5")
	if err != nil {
		t.Fatal(err)
	}
	after, captured, err := before.Apply(move)
	if err != nil {
		t.Fatal(err)
	}
	match := Match{FEN: before.FEN()}

	appendMove(&match, before, after, move, captured, "player", 0)

	if len(match.Moves) != 1 || !match.Moves[0].GivesCheck {
		t.Fatalf("move record must mark a checking move: %+v", match.Moves)
	}
}
