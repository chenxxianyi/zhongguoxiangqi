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
	if !xiangqi.InitialPosition().IsLegal(result.BestMove) {
		t.Fatal("cancelled analysis must still return a safe legal fallback")
	}
}
