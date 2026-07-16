package game

import (
	"context"
	"testing"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/engine"
)

type firstMoveEngine struct {
	delay time.Duration
}

func (e firstMoveEngine) Name() string                 { return "test-first-move" }
func (e firstMoveEngine) Health(context.Context) error { return nil }
func (e firstMoveEngine) Close() error                 { return nil }
func (e firstMoveEngine) Analyze(ctx context.Context, request engine.AnalyzeRequest) (engine.AnalyzeResult, error) {
	if e.delay > 0 {
		select {
		case <-time.After(e.delay):
		case <-ctx.Done():
		}
	}
	move := request.Position.LegalMoves()[0]
	return engine.AnalyzeResult{BestMove: move, BestMoveICCS: move.ICCS()}, nil
}

type fixedBook struct {
	move string
}

func (b fixedBook) SelectBookMove(_ xiangqi.Position, _ string) (xiangqi.Move, bool) {
	move, err := xiangqi.ParseMove(b.move)
	return move, err == nil
}

func TestCreateStoresAIModeAndUsesBookAdvisor(t *testing.T) {
	service := NewService(NewMemoryRepository(), firstMoveEngine{}, NewEventBus(), time.Second)
	service.SetBookAdvisor(fixedBook{move: "a6a5"})
	match, err := service.Create(CreateRequest{PlayerColor: "red", Difficulty: 1, AIMode: "library"}, "create-book")
	if err != nil {
		t.Fatal(err)
	}
	if match.AIMode != AIModeLibrary {
		t.Fatalf("ai mode = %s", match.AIMode)
	}
	if _, err := service.Create(CreateRequest{PlayerColor: "red", Difficulty: 1, AIMode: "unknown"}, ""); err == nil {
		t.Fatal("expected invalid aiMode error")
	}
	if _, err := service.ApplyPlayerMove(match.ID, MoveRequest{Move: "a3a4", ExpectedMatchVersion: match.Version}, "move-book"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		current, err := service.Get(match.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(current.Moves) == 2 {
			if current.Moves[1].Move != "a6a5" {
				t.Fatalf("book move not used: %+v", current.Moves)
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("book move was not applied")
}

func TestMatchMoveVersionAndAsyncAI(t *testing.T) {
	service := NewService(NewMemoryRepository(), firstMoveEngine{}, NewEventBus(), time.Second)
	match, err := service.Create(CreateRequest{PlayerColor: "red", Difficulty: 1}, "create-1")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := service.ApplyPlayerMove(match.ID, MoveRequest{
		Move: "a3a4", ExpectedMatchVersion: match.Version,
	}, "move-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != StatusAIThinking {
		t.Fatalf("status = %s", updated.Status)
	}
	if _, err := service.ApplyPlayerMove(match.ID, MoveRequest{
		Move: "c3c4", ExpectedMatchVersion: match.Version,
	}, "move-stale"); err != ErrVersionConflict {
		t.Fatalf("stale move error = %v", err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		current, err := service.Get(match.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(current.Moves) == 2 {
			if current.Status != StatusPlayerTurn || current.SideToMove != "red" {
				t.Fatalf("unexpected snapshot: %+v", current)
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("AI move was not applied")
}

func TestUndoCancelsStaleAIResult(t *testing.T) {
	service := NewService(NewMemoryRepository(), firstMoveEngine{delay: 80 * time.Millisecond}, NewEventBus(), time.Second)
	match, _ := service.Create(CreateRequest{PlayerColor: "red", Difficulty: 1}, "")
	afterMove, err := service.ApplyPlayerMove(match.ID, MoveRequest{Move: "a3a4", ExpectedMatchVersion: 1}, "")
	if err != nil {
		t.Fatal(err)
	}
	undone, err := service.Undo(match.ID, VersionRequest{ExpectedMatchVersion: afterMove.Version}, "")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(150 * time.Millisecond)
	current, _ := service.Get(match.ID)
	if len(current.Moves) != 0 || current.FEN != xiangqi.InitialFEN || current.Version != undone.Version {
		t.Fatalf("stale AI result polluted undone match: %+v", current)
	}
}

func TestIdempotentMoveReturnsCurrentMatch(t *testing.T) {
	service := NewService(NewMemoryRepository(), firstMoveEngine{delay: time.Second}, NewEventBus(), 2*time.Second)
	match, _ := service.Create(CreateRequest{PlayerColor: "red", Difficulty: 1}, "")
	request := MoveRequest{Move: "a3a4", ExpectedMatchVersion: match.Version}
	first, err := service.ApplyPlayerMove(match.ID, request, "same-key")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.ApplyPlayerMove(match.ID, request, "same-key")
	if err != nil {
		t.Fatal(err)
	}
	if second.Version != first.Version || len(second.Moves) != 1 {
		t.Fatalf("idempotent request applied twice: %+v", second)
	}
}
