package xiangqi

import "testing"

func TestInitialFENRoundTrip(t *testing.T) {
	position, err := ParseFEN(InitialFEN)
	if err != nil {
		t.Fatal(err)
	}
	if got := position.FEN(); got != InitialFEN {
		t.Fatalf("FEN round-trip mismatch:\nwant %s\n got %s", InitialFEN, got)
	}
	if position.SideToMove() != Red {
		t.Fatalf("initial side = %v, want red", position.SideToMove())
	}
	if len(position.LegalMoves()) == 0 {
		t.Fatal("initial position must have legal moves")
	}
}

func TestSquareAndMoveICCSRoundTrip(t *testing.T) {
	for _, raw := range []string{"a0", "i9", "e4"} {
		square, err := ParseSquare(raw)
		if err != nil {
			t.Fatal(err)
		}
		if got := square.ICCS(); got != raw {
			t.Fatalf("square round-trip: want %s, got %s", raw, got)
		}
	}
	move, err := ParseMove("a0a1")
	if err != nil {
		t.Fatal(err)
	}
	if move.ICCS() != "a0a1" {
		t.Fatalf("move round-trip: got %s", move.ICCS())
	}
}

func TestHorseLeg(t *testing.T) {
	clear := mustPosition(t, "3k5/9/9/9/9/9/9/9/4H4/4K4 w")
	assertLegal(t, clear, "e1g2", true)

	blocked := mustPosition(t, "3k5/9/9/9/9/9/9/9/5P3/4K4 w")
	// Put the horse and its right leg blocker on e1/f1.
	blocked.board[8][4] = Piece{Color: Red, Type: Horse}
	assertLegal(t, blocked, "e1g2", false)
}

func TestElephantEyeAndRiver(t *testing.T) {
	position := mustPosition(t, "3k5/9/9/9/9/9/9/9/2E6/4K4 w")
	assertLegal(t, position, "c1e3", true)
	assertLegal(t, position, "c1a3", true)

	blocked := position
	blocked.board[7][3] = Piece{Color: Red, Type: Pawn}
	assertLegal(t, blocked, "c1e3", false)

	river := mustPosition(t, "3k5/9/9/9/2E6/9/9/9/9/4K4 w")
	assertLegal(t, river, "c5e7", false)
}

func TestCannonScreen(t *testing.T) {
	noScreen := mustPosition(t, "3k5/9/9/9/9/9/9/4r4/9/C3K4 w")
	assertLegal(t, noScreen, "a0e2", false)

	oneScreen := mustPosition(t, "3k5/9/9/9/9/9/9/4r4/P8/C3K4 w")
	assertLegal(t, oneScreen, "a0a7", false) // destination empty beyond screen

	capture := mustPosition(t, "r2k5/9/9/9/9/9/9/9/P8/C3K4 w")
	assertLegal(t, capture, "a0a9", true)

	twoScreens := mustPosition(t, "r2k5/9/9/9/9/9/P8/9/P8/C3K4 w")
	assertLegal(t, twoScreens, "a0a9", false)
}

func TestPalaceAndPawnRiverRestrictions(t *testing.T) {
	general := mustPosition(t, "3k5/9/9/9/9/9/9/4K4/9/9 w")
	assertLegal(t, general, "e2f2", true)
	assertLegal(t, general, "e2e1", true)
	assertLegal(t, general, "e2e3", false)

	pawnBefore := mustPosition(t, "3k5/9/9/9/9/9/4P4/9/9/4K4 w")
	assertLegal(t, pawnBefore, "e3d3", false)
	assertLegal(t, pawnBefore, "e3e4", true)
	assertLegal(t, pawnBefore, "e3e2", false)

	pawnAfter := mustPosition(t, "3k5/9/9/9/4P4/9/9/9/9/4K4 w")
	assertLegal(t, pawnAfter, "e5d5", true)
}

func TestGeneralsFacingAndSelfCheck(t *testing.T) {
	facing := mustPosition(t, "4k4/9/9/9/4R4/9/9/9/9/4K4 w")
	assertLegal(t, facing, "e5d5", false)
	assertLegal(t, facing, "e5e6", true)

	pinned := mustPosition(t, "4k4/9/9/9/9/9/4R4/9/9/4K4 w")
	assertLegal(t, pinned, "e3f3", false)
}

func TestCheckmateAndNoLegalMove(t *testing.T) {
	checkmate := mustPosition(t, "4k4/3RR4/6H2/9/9/9/9/9/9/4K4 b")
	mateResult := checkmate.Adjudicate()
	if mateResult.Outcome != OutcomeRedWin || mateResult.Termination != TerminationCheckmate || !mateResult.InCheck {
		t.Fatalf("checkmate adjudication: %+v", mateResult)
	}

	noMoves := mustPosition(t, "4k4/3R1R3/9/9/4P4/9/9/9/9/4K4 b")
	noMovesResult := noMoves.Adjudicate()
	if noMovesResult.Outcome != OutcomeRedWin || noMovesResult.Termination != TerminationNoMoves || noMovesResult.InCheck {
		t.Fatalf("no-legal-move adjudication: %+v", noMovesResult)
	}
}

func TestApplyDoesNotMutateOriginalAndHashIsStable(t *testing.T) {
	position := InitialPosition()
	beforeFEN, beforeHash := position.FEN(), position.Hash()
	move, _ := ParseMove("a3a4")
	next, _, err := position.Apply(move)
	if err != nil {
		t.Fatal(err)
	}
	if position.FEN() != beforeFEN || position.Hash() != beforeHash {
		t.Fatal("immutable Apply mutated original position")
	}
	if next.FEN() == beforeFEN || next.Hash() == beforeHash {
		t.Fatal("applied position did not change")
	}
	again, err := ParseFEN(beforeFEN)
	if err != nil {
		t.Fatal(err)
	}
	if again.Hash() != beforeHash {
		t.Fatal("hash is not reproducible")
	}
}

func TestMalformedFENRejected(t *testing.T) {
	cases := []string{
		"",
		"9/9 w",
		"9/9/9/9/9/9/9/9/9/9 w",
		"4k4/9/9/9/9/9/9/9/9/4X4 w",
		"4k4/9/9/9/9/9/9/9/9/4K4 x",
	}
	for _, raw := range cases {
		if _, err := ParseFEN(raw); err == nil {
			t.Fatalf("expected malformed FEN to fail: %q", raw)
		}
	}
}

func mustPosition(t *testing.T, fen string) Position {
	t.Helper()
	position, err := ParseFEN(fen)
	if err != nil {
		t.Fatal(err)
	}
	return position
}

func assertLegal(t *testing.T, position Position, raw string, want bool) {
	t.Helper()
	move, err := ParseMove(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := position.IsLegal(move); got != want {
		t.Fatalf("IsLegal(%s) = %t, want %t for %s", raw, got, want, position.FEN())
	}
}
