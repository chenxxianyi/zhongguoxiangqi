package game

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/engine"
	"xiangqi-lab/internal/engine/difficulty"
)

type Service struct {
	repository  Repository
	engine      engine.Engine
	events      *EventBus
	moveTimeCap time.Duration
	book        BookAdvisor

	cancelMu sync.Mutex
	searches map[string]context.CancelFunc
}

type BookAdvisor interface {
	SelectBookMove(position xiangqi.Position, mode string) (xiangqi.Move, bool)
}

type CreateRequest struct {
	PlayerColor string `json:"playerColor"`
	Difficulty  int    `json:"difficulty"`
	AIMode      string `json:"aiMode,omitempty"`
	InitialFEN  string `json:"initialFen,omitempty"`
	AllowUndo   *bool  `json:"allowUndo,omitempty"`
}

type MoveRequest struct {
	Move                 string `json:"move"`
	ExpectedMatchVersion int64  `json:"expectedMatchVersion"`
}

type VersionRequest struct {
	ExpectedMatchVersion int64 `json:"expectedMatchVersion"`
}

func NewService(repository Repository, searchEngine engine.Engine, events *EventBus, moveTimeCap time.Duration) *Service {
	return &Service{
		repository: repository, engine: searchEngine, events: events,
		moveTimeCap: moveTimeCap, searches: make(map[string]context.CancelFunc),
	}
}

func (s *Service) AuthoritativeStore() string {
	return s.repository.Name()
}

func (s *Service) SetBookAdvisor(book BookAdvisor) {
	s.book = book
}

func (s *Service) Create(request CreateRequest, idempotencyKey string) (Snapshot, error) {
	digest := requestDigest(request)
	if existing, ok, err := s.repository.Idempotency("create-match", idempotencyKey, digest); err != nil {
		return Snapshot{}, err
	} else if ok {
		return existing.Snapshot(), nil
	}
	player, err := parseColor(request.PlayerColor)
	if err != nil {
		return Snapshot{}, err
	}
	if request.Difficulty < 1 || request.Difficulty > 10 {
		return Snapshot{}, fmt.Errorf("difficulty must be between 1 and 10")
	}
	aiMode, err := parseAIMode(request.AIMode)
	if err != nil {
		return Snapshot{}, err
	}
	position := xiangqi.InitialPosition()
	if strings.TrimSpace(request.InitialFEN) != "" {
		position, err = xiangqi.ParseFEN(request.InitialFEN)
		if err != nil {
			return Snapshot{}, fmt.Errorf("initial FEN: %w", err)
		}
	}
	allowUndo := true
	if request.AllowUndo != nil {
		allowUndo = *request.AllowUndo
	}
	now := time.Now().UTC()
	match := Match{
		ID: newID(), Version: 1, PlayerColor: player, Difficulty: request.Difficulty, AIMode: aiMode,
		Engine: s.engine.Name(), AllowUndo: allowUndo, InitialFEN: position.FEN(), FEN: position.FEN(),
		SideToMove: position.SideToMove(), Outcome: xiangqi.OutcomeOngoing,
		CreatedAt: now, UpdatedAt: now, Moves: []MoveRecord{},
	}
	if position.SideToMove() == player {
		match.Status = StatusPlayerTurn
	} else {
		match.Status = StatusAIThinking
	}
	if adjudication := position.Adjudicate(); adjudication.Outcome != xiangqi.OutcomeOngoing {
		match.Status, match.Outcome, match.Termination = StatusFinished, adjudication.Outcome, adjudication.Termination
	}
	if err := s.repository.Create(match); err != nil {
		return Snapshot{}, err
	}
	s.repository.SaveIdempotency("create-match", idempotencyKey, digest, match)
	s.events.Publish(match.ID, match.Version, "match.snapshot", match.Snapshot())
	if match.Status == StatusAIThinking {
		s.scheduleAI(match.ID, match.Version)
	}
	return match.Snapshot(), nil
}

func (s *Service) Get(id string) (Snapshot, error) {
	match, err := s.repository.Get(id)
	if err != nil {
		return Snapshot{}, err
	}
	return match.Snapshot(), nil
}

func (s *Service) LegalMoves(id, from string) (LegalMovesResponse, error) {
	match, err := s.repository.Get(id)
	if err != nil {
		return LegalMovesResponse{}, err
	}
	response := LegalMovesResponse{
		MatchID: match.ID, MatchVersion: match.Version,
		SideToMove: match.SideToMove.String(), Moves: []LegalMove{},
	}
	if match.Status != StatusPlayerTurn || match.SideToMove != match.PlayerColor {
		return response, nil
	}
	position, err := xiangqi.ParseFEN(match.FEN)
	if err != nil {
		return LegalMovesResponse{}, err
	}
	var fromSquare *xiangqi.Square
	if strings.TrimSpace(from) != "" {
		square, err := xiangqi.ParseSquare(from)
		if err != nil {
			return LegalMovesResponse{}, fmt.Errorf("invalid from square: %w", err)
		}
		fromSquare = &square
	}
	for _, move := range position.LegalMoves() {
		if fromSquare != nil && move.From != *fromSquare {
			continue
		}
		response.Moves = append(response.Moves, LegalMove{
			Move: move.ICCS(), From: move.From.ICCS(), To: move.To.ICCS(),
			Capture: !position.PieceAt(move.To).Empty(),
		})
	}
	return response, nil
}

func (s *Service) List() []Snapshot {
	matches := s.repository.List()
	sort.Slice(matches, func(i, j int) bool { return matches[i].CreatedAt.After(matches[j].CreatedAt) })
	result := make([]Snapshot, 0, len(matches))
	for _, match := range matches {
		result = append(result, match.Snapshot())
	}
	return result
}

func (s *Service) ApplyPlayerMove(id string, request MoveRequest, idempotencyKey string) (Snapshot, error) {
	digest := requestDigest(request)
	route := "move|" + id
	if existing, ok, err := s.repository.Idempotency(route, idempotencyKey, digest); err != nil {
		return Snapshot{}, err
	} else if ok {
		return existing.Snapshot(), nil
	}
	move, err := xiangqi.ParseMove(request.Move)
	if err != nil {
		return Snapshot{}, err
	}
	updated, err := s.repository.Update(id, request.ExpectedMatchVersion, func(match *Match) error {
		if match.Status != StatusPlayerTurn || match.SideToMove != match.PlayerColor {
			return ErrNotPlayerTurn
		}
		position, err := xiangqi.ParseFEN(match.FEN)
		if err != nil {
			return err
		}
		if !position.IsLegal(move) {
			return xiangqi.ErrIllegalMove
		}
		next, captured, err := position.Apply(move)
		if err != nil {
			return err
		}
		appendMove(match, position, next, move, captured, "player", 0)
		match.Version++
		match.DrawOffered = false
		applyAdjudication(match, next)
		if match.Status != StatusFinished {
			match.Status = StatusAIThinking
		}
		return nil
	})
	if err != nil {
		return Snapshot{}, err
	}
	s.repository.SaveIdempotency(route, idempotencyKey, digest, updated)
	s.events.Publish(id, updated.Version, "match.move_accepted", updated.Moves[len(updated.Moves)-1])
	if updated.Status == StatusFinished {
		s.events.Publish(id, updated.Version, "match.finished", updated.Snapshot())
	} else {
		s.events.Publish(id, updated.Version, "match.ai_thinking", map[string]any{"engine": updated.Engine})
		s.scheduleAI(id, updated.Version)
	}
	return updated.Snapshot(), nil
}

func (s *Service) Undo(id string, request VersionRequest, idempotencyKey string) (Snapshot, error) {
	s.cancelSearch(id)
	digest := requestDigest(request)
	route := "undo|" + id
	if existing, ok, err := s.repository.Idempotency(route, idempotencyKey, digest); err != nil {
		return Snapshot{}, err
	} else if ok {
		return existing.Snapshot(), nil
	}
	updated, err := s.repository.Update(id, request.ExpectedMatchVersion, func(match *Match) error {
		if !match.AllowUndo {
			return ErrUndoDisabled
		}
		if match.Status == StatusFinished || match.Status == StatusAborted {
			return ErrStateConflict
		}
		if len(match.Moves) == 0 {
			return ErrNoMovesToUndo
		}
		remove := 1
		if match.Status != StatusAIThinking && len(match.Moves) >= 2 {
			remove = 2
		}
		match.Moves = append([]MoveRecord(nil), match.Moves[:len(match.Moves)-remove]...)
		position, err := replay(match.InitialFEN, match.Moves)
		if err != nil {
			return err
		}
		match.FEN, match.SideToMove = position.FEN(), position.SideToMove()
		match.Version++
		match.Outcome, match.Termination, match.DrawOffered = xiangqi.OutcomeOngoing, xiangqi.TerminationNone, false
		match.UpdatedAt = time.Now().UTC()
		if match.SideToMove == match.PlayerColor {
			match.Status = StatusPlayerTurn
		} else {
			match.Status = StatusAIThinking
		}
		return nil
	})
	if err != nil {
		return Snapshot{}, err
	}
	s.repository.SaveIdempotency(route, idempotencyKey, digest, updated)
	s.events.Publish(id, updated.Version, "match.undo_applied", updated.Snapshot())
	if updated.Status == StatusAIThinking {
		s.scheduleAI(id, updated.Version)
	}
	return updated.Snapshot(), nil
}

func (s *Service) Resign(id string, request VersionRequest, idempotencyKey string) (Snapshot, error) {
	s.cancelSearch(id)
	digest := requestDigest(request)
	route := "resign|" + id
	if existing, ok, err := s.repository.Idempotency(route, idempotencyKey, digest); err != nil {
		return Snapshot{}, err
	} else if ok {
		return existing.Snapshot(), nil
	}
	updated, err := s.repository.Update(id, request.ExpectedMatchVersion, func(match *Match) error {
		if match.Status == StatusFinished || match.Status == StatusAborted {
			return ErrStateConflict
		}
		match.Version++
		match.Status = StatusFinished
		match.Termination = xiangqi.TerminationResign
		if match.PlayerColor == xiangqi.Red {
			match.Outcome = xiangqi.OutcomeBlackWin
		} else {
			match.Outcome = xiangqi.OutcomeRedWin
		}
		match.UpdatedAt = time.Now().UTC()
		return nil
	})
	if err != nil {
		return Snapshot{}, err
	}
	s.repository.SaveIdempotency(route, idempotencyKey, digest, updated)
	s.events.Publish(id, updated.Version, "match.finished", updated.Snapshot())
	return updated.Snapshot(), nil
}

func (s *Service) OfferDraw(id string, request VersionRequest, idempotencyKey string) (Snapshot, bool, error) {
	digest := requestDigest(request)
	route := "draw|" + id
	if existing, ok, err := s.repository.Idempotency(route, idempotencyKey, digest); err != nil {
		return Snapshot{}, false, err
	} else if ok {
		return existing.Snapshot(), existing.Outcome == xiangqi.OutcomeDraw, nil
	}
	accepted := false
	updated, err := s.repository.Update(id, request.ExpectedMatchVersion, func(match *Match) error {
		if match.Status == StatusFinished || match.Status == StatusAborted {
			return ErrStateConflict
		}
		// A deterministic, disclosed MVP policy: the built-in opponent accepts
		// after 40 plies when material is within one minor piece.
		position, err := xiangqi.ParseFEN(match.FEN)
		if err != nil {
			return err
		}
		accepted = len(match.Moves) >= 40 && abs(materialBalance(position)) <= 450
		match.Version++
		match.DrawOffered = !accepted
		match.UpdatedAt = time.Now().UTC()
		if accepted {
			match.Status, match.Outcome, match.Termination = StatusFinished, xiangqi.OutcomeDraw, xiangqi.TerminationAgreement
		}
		return nil
	})
	if err != nil {
		return Snapshot{}, false, err
	}
	s.repository.SaveIdempotency(route, idempotencyKey, digest, updated)
	eventType := "match.draw_declined"
	if accepted {
		eventType = "match.finished"
	}
	s.events.Publish(id, updated.Version, eventType, updated.Snapshot())
	return updated.Snapshot(), accepted, nil
}

func (s *Service) scheduleAI(id string, expectedVersion int64) {
	s.cancelSearch(id)
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelMu.Lock()
	s.searches[id] = cancel
	s.cancelMu.Unlock()
	go s.runAI(ctx, id, expectedVersion)
}

func (s *Service) runAI(ctx context.Context, id string, expectedVersion int64) {
	started := time.Now()
	match, err := s.repository.Get(id)
	if err != nil || match.Version != expectedVersion || match.Status != StatusAIThinking {
		return
	}
	position, err := xiangqi.ParseFEN(match.FEN)
	if err != nil {
		s.markRecoverable(id, expectedVersion, "invalid_authoritative_fen")
		return
	}
	if s.book != nil && match.AIMode != AIModeStandard {
		if bookMove, ok := s.book.SelectBookMove(position, string(match.AIMode)); ok && position.IsLegal(bookMove) {
			if s.applyAIMove(id, expectedVersion, bookMove, started, 0, 0, "learning_book") {
				return
			}
		}
	}
	profile := difficulty.Get(match.Difficulty)
	moveTime := profile.MoveTime
	if s.moveTimeCap > 0 && moveTime > s.moveTimeCap {
		moveTime = s.moveTimeCap
	}
	result, err := s.engine.Analyze(ctx, engine.AnalyzeRequest{
		Position: position, MaxDepth: profile.MaxDepth, MaxNodes: profile.MaxNodes,
		MoveTime: moveTime, MultiPV: profile.MultiPV,
	})
	if err != nil {
		if !errors.Is(ctx.Err(), context.Canceled) {
			s.markRecoverable(id, expectedVersion, "engine_failed")
		}
		return
	}
	if !position.IsLegal(result.BestMove) {
		s.markRecoverable(id, expectedVersion, "engine_returned_illegal_move")
		return
	}
	s.applyAIMove(id, expectedVersion, result.BestMove, started, result.Depth, result.Nodes, result.StoppedReason)
}

func (s *Service) applyAIMove(
	id string,
	expectedVersion int64,
	move xiangqi.Move,
	started time.Time,
	depth int,
	nodes uint64,
	stoppedReason string,
) bool {
	updated, err := s.repository.Update(id, expectedVersion, func(current *Match) error {
		if current.Status != StatusAIThinking || current.SideToMove == current.PlayerColor {
			return ErrStateConflict
		}
		authoritative, err := xiangqi.ParseFEN(current.FEN)
		if err != nil {
			return err
		}
		if !authoritative.IsLegal(move) {
			return xiangqi.ErrIllegalMove
		}
		next, captured, err := authoritative.Apply(move)
		if err != nil {
			return err
		}
		appendMove(current, authoritative, next, move, captured, "ai", time.Since(started))
		current.Version++
		applyAdjudication(current, next)
		if current.Status != StatusFinished {
			current.Status = StatusPlayerTurn
		}
		return nil
	})
	if err != nil {
		return false // stale/cancelled results are intentionally discarded
	}
	s.events.Publish(id, updated.Version, "match.ai_move_applied", map[string]any{
		"move": updated.Moves[len(updated.Moves)-1], "depth": depth,
		"nodes": nodes, "stoppedReason": stoppedReason,
	})
	if updated.Status == StatusFinished {
		s.events.Publish(id, updated.Version, "match.finished", updated.Snapshot())
	}
	return true
}

func (s *Service) markRecoverable(id string, expectedVersion int64, reason string) {
	updated, err := s.repository.Update(id, expectedVersion, func(match *Match) error {
		match.Version++
		match.Status = StatusRecoverableError
		match.UpdatedAt = time.Now().UTC()
		return nil
	})
	if err == nil {
		s.events.Publish(id, updated.Version, "match.engine_degraded", map[string]string{"reason": reason})
	}
}

func (s *Service) cancelSearch(id string) {
	s.cancelMu.Lock()
	cancel := s.searches[id]
	delete(s.searches, id)
	s.cancelMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func appendMove(match *Match, before, after xiangqi.Position, move xiangqi.Move, captured xiangqi.Piece, actor string, thinkTime time.Duration) {
	match.Moves = append(match.Moves, MoveRecord{
		Ply: len(match.Moves) + 1, Move: move.ICCS(), Side: before.SideToMove().String(),
		Actor: actor, Captured: pieceName(captured), FENBefore: before.FEN(), FENAfter: after.FEN(),
		HashAfter: fmt.Sprintf("%016x", after.Hash()), PlayedAt: time.Now().UTC(),
		ThinkTimeMs: thinkTime.Milliseconds(),
	})
	match.FEN, match.SideToMove, match.UpdatedAt = after.FEN(), after.SideToMove(), time.Now().UTC()
}

func applyAdjudication(match *Match, position xiangqi.Position) {
	adjudication := position.Adjudicate()
	if adjudication.Outcome != xiangqi.OutcomeOngoing {
		match.Status, match.Outcome, match.Termination = StatusFinished, adjudication.Outcome, adjudication.Termination
		return
	}
	hash := fmt.Sprintf("%016x", position.Hash())
	count := 0
	initial, _ := xiangqi.ParseFEN(match.InitialFEN)
	if fmt.Sprintf("%016x", initial.Hash()) == hash {
		count++
	}
	for _, move := range match.Moves {
		if move.HashAfter == hash {
			count++
		}
	}
	if count >= 3 {
		match.Status, match.Outcome, match.Termination = StatusFinished, xiangqi.OutcomeDraw, xiangqi.TerminationRepetition
	}
}

func replay(initialFEN string, moves []MoveRecord) (xiangqi.Position, error) {
	position, err := xiangqi.ParseFEN(initialFEN)
	if err != nil {
		return xiangqi.Position{}, err
	}
	for _, record := range moves {
		move, err := xiangqi.ParseMove(record.Move)
		if err != nil {
			return xiangqi.Position{}, err
		}
		position, _, err = position.Apply(move)
		if err != nil {
			return xiangqi.Position{}, err
		}
	}
	return position, nil
}

func parseColor(raw string) (xiangqi.Color, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "red":
		return xiangqi.Red, nil
	case "black":
		return xiangqi.Black, nil
	default:
		return xiangqi.NoColor, fmt.Errorf("playerColor must be red or black")
	}
}

func parseAIMode(raw string) (AIMode, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(AIModeStandard):
		return AIModeStandard, nil
	case string(AIModeLibrary):
		return AIModeLibrary, nil
	case string(AIModeStyle):
		return AIModeStyle, nil
	default:
		return "", fmt.Errorf("aiMode must be standard, library or style")
	}
}

func requestDigest(value any) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%#v", value)))
	return hex.EncodeToString(sum[:])
}

func materialBalance(position xiangqi.Position) int {
	values := map[xiangqi.PieceType]int{
		xiangqi.Rook: 900, xiangqi.Cannon: 450, xiangqi.Horse: 420,
		xiangqi.Elephant: 200, xiangqi.Advisor: 200, xiangqi.Pawn: 100,
	}
	score := 0
	for rank := 0; rank < xiangqi.Ranks; rank++ {
		for file := 0; file < xiangqi.Files; file++ {
			piece := position.PieceAt(xiangqi.Square{File: file, Rank: rank})
			if piece.Color == xiangqi.Red {
				score += values[piece.Type]
			} else if piece.Color == xiangqi.Black {
				score -= values[piece.Type]
			}
		}
	}
	return score
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
