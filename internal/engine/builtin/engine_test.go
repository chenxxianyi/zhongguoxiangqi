package builtin

import (
	"context"
	"testing"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/engine"
)

func TestAnalyzeReturnsLegalMove(t *testing.T) {
	position := xiangqi.InitialPosition()
	result, err := New().Analyze(context.Background(), engine.AnalyzeRequest{
		Position: position, MaxDepth: 2, MaxNodes: 5_000, MoveTime: time.Second, MultiPV: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !position.IsLegal(result.BestMove) {
		t.Fatalf("engine returned illegal move %s", result.BestMoveICCS)
	}
	if len(result.Candidates) == 0 || result.Nodes == 0 {
		t.Fatalf("incomplete result: %+v", result)
	}
}

func TestAnalyzeHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := New().Analyze(ctx, engine.AnalyzeRequest{
		Position: xiangqi.InitialPosition(), MaxDepth: 8, MaxNodes: 10_000_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.StoppedReason != "time_or_context" {
		t.Fatalf("stopped reason = %q", result.StoppedReason)
	}
	position := xiangqi.InitialPosition()
	if !position.IsLegal(result.BestMove) {
		t.Fatal("cancelled analysis must still return a safe legal fallback")
	}
}

func TestAnalyzeAvoidsPoisonedOpeningCannonCapture(t *testing.T) {
	position := xiangqi.InitialPosition()
	result, err := New().Analyze(context.Background(), engine.AnalyzeRequest{
		Position: position, MaxDepth: 1, MaxNodes: 100_000, MultiPV: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.BestMoveICCS == "b2b9" || result.BestMoveICCS == "h2h9" {
		t.Fatalf("engine chose poisoned cannon capture %s at depth 1", result.BestMoveICCS)
	}
	if result.Depth != 1 {
		t.Fatalf("completed depth = %d, want 1", result.Depth)
	}
}

func TestMoveOrderingPrefersHighValueCapture(t *testing.T) {
	position, err := xiangqi.ParseFEN("4k4/9/9/9/p2Pr4/R8/9/9/9/4K4 w")
	if err != nil {
		t.Fatal(err)
	}
	state := &searchState{ctx: context.Background(), maxNodes: 1_000, table: makeTranspositionTable(1_000)}
	moves := orderedMoves(state, &position, 0, xiangqi.Move{})
	if len(moves) == 0 {
		t.Fatal("expected legal moves")
	}
	if moves[0].ICCS() != "d5e5" {
		t.Fatalf("first move = %s, want pawn captures rook d5e5", moves[0].ICCS())
	}
}

func TestTranspositionTableKeepsDeeperEntry(t *testing.T) {
	state := &searchState{table: make([]ttEntry, 4)}
	move, err := xiangqi.ParseMove("a3a4")
	if err != nil {
		t.Fatal(err)
	}
	state.store(7, 5, 42, ttExact, move)
	state.store(7, 3, 99, ttLower, move)
	entry, found := state.probe(7)
	if !found {
		t.Fatal("expected transposition table hit")
	}
	if entry.depth != 5 || entry.score != 42 || entry.bound != ttExact {
		t.Fatalf("deeper entry was replaced: %+v", entry)
	}

	mate := mateScore - 12
	if restored := scoreFromTT(scoreToTT(mate, 6), 6); restored != mate {
		t.Fatalf("mate score restored as %d, want %d", restored, mate)
	}
}

var benchmarkResult engine.AnalyzeResult

func BenchmarkAnalyzeInitialPosition(b *testing.B) {
	searchEngine := New()
	position := xiangqi.InitialPosition()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := searchEngine.Analyze(context.Background(), engine.AnalyzeRequest{
			Position: position, MaxDepth: 10, MaxNodes: 25_000, MultiPV: 1,
		})
		if err != nil {
			b.Fatal(err)
		}
		benchmarkResult = result
	}
}
