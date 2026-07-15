package analysis

import (
	"context"
	"testing"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/engine"
	"xiangqi-lab/internal/game"
)

type matchSource struct{ snapshot game.Snapshot }

func (m matchSource) Get(string) (game.Snapshot, error) { return m.snapshot, nil }

type deterministicEngine struct{}

func (deterministicEngine) Name() string                 { return "deterministic" }
func (deterministicEngine) Health(context.Context) error { return nil }
func (deterministicEngine) Close() error                 { return nil }
func (deterministicEngine) Analyze(_ context.Context, request engine.AnalyzeRequest) (engine.AnalyzeResult, error) {
	move := request.Position.LegalMoves()[0]
	return engine.AnalyzeResult{
		BestMove: move, BestMoveICCS: move.ICCS(), Depth: 1, Nodes: 1,
		Candidates: []engine.Candidate{{Move: move, MoveICCS: move.ICCS()}},
	}, nil
}

func TestAnalysisJobProducesBoundedResult(t *testing.T) {
	position := xiangqi.InitialPosition()
	move := position.LegalMoves()[0]
	next, _, _ := position.Apply(move)
	snapshot := game.Snapshot{
		ID: "match-1", InitialFEN: position.FEN(),
		Moves: []game.MoveRecord{{
			Ply: 1, Move: move.ICCS(), Side: "red",
			FENBefore: position.FEN(), FENAfter: next.FEN(),
		}},
	}
	service := NewService(matchSource{snapshot}, deterministicEngine{})
	job, err := service.CreateJob(CreateJobRequest{MatchID: snapshot.ID})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		current, _ := service.GetJob(job.ID)
		if current.Status == "completed" {
			result, err := service.GetResult(snapshot.ID)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Moves) != 1 || result.Moves[0].Classification != "best" {
				t.Fatalf("result: %+v", result)
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("analysis timeout")
}
