package game

import (
	"fmt"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
)

type Status string

const (
	StatusPlayerTurn       Status = "active_player_turn"
	StatusAIThinking       Status = "active_ai_thinking"
	StatusFinished         Status = "finished"
	StatusAborted          Status = "aborted"
	StatusRecoverableError Status = "recoverable_error"
)

type AIMode string

const (
	AIModeStandard AIMode = "standard"
	AIModeLibrary  AIMode = "library"
	AIModeStyle    AIMode = "style"
)

type MoveRecord struct {
	Ply         int       `json:"ply"`
	Move        string    `json:"move"`
	Side        string    `json:"side"`
	Actor       string    `json:"actor"`
	Captured    string    `json:"captured,omitempty"`
	FENBefore   string    `json:"fenBefore"`
	FENAfter    string    `json:"fenAfter"`
	HashAfter   string    `json:"hashAfter"`
	PlayedAt    time.Time `json:"playedAt"`
	ThinkTimeMs int64     `json:"thinkTimeMs,omitempty"`
}

type Match struct {
	ID          string              `json:"id"`
	Version     int64               `json:"version"`
	Status      Status              `json:"status"`
	PlayerColor xiangqi.Color       `json:"-"`
	Difficulty  int                 `json:"difficulty"`
	AIMode      AIMode              `json:"aiMode"`
	Engine      string              `json:"engine"`
	AllowUndo   bool                `json:"allowUndo"`
	InitialFEN  string              `json:"initialFen"`
	FEN         string              `json:"fen"`
	SideToMove  xiangqi.Color       `json:"-"`
	Moves       []MoveRecord        `json:"moves"`
	Outcome     xiangqi.Outcome     `json:"outcome"`
	Termination xiangqi.Termination `json:"termination,omitempty"`
	DrawOffered bool                `json:"drawOffered"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

type Snapshot struct {
	ID          string              `json:"id"`
	Version     int64               `json:"version"`
	Status      Status              `json:"status"`
	PlayerColor string              `json:"playerColor"`
	SideToMove  string              `json:"sideToMove"`
	Difficulty  int                 `json:"difficulty"`
	AIMode      AIMode              `json:"aiMode"`
	Engine      string              `json:"engine"`
	AllowUndo   bool                `json:"allowUndo"`
	InitialFEN  string              `json:"initialFen"`
	FEN         string              `json:"fen"`
	Moves       []MoveRecord        `json:"moves"`
	Outcome     xiangqi.Outcome     `json:"outcome"`
	Termination xiangqi.Termination `json:"termination,omitempty"`
	DrawOffered bool                `json:"drawOffered"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

func (m Match) Snapshot() Snapshot {
	moves := append([]MoveRecord(nil), m.Moves...)
	return Snapshot{
		ID: m.ID, Version: m.Version, Status: m.Status,
		PlayerColor: m.PlayerColor.String(), SideToMove: m.SideToMove.String(),
		Difficulty: m.Difficulty, AIMode: m.AIMode, Engine: m.Engine, AllowUndo: m.AllowUndo,
		InitialFEN: m.InitialFEN, FEN: m.FEN, Moves: moves,
		Outcome: m.Outcome, Termination: m.Termination, DrawOffered: m.DrawOffered,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func pieceName(piece xiangqi.Piece) string {
	if piece.Empty() {
		return ""
	}
	names := map[xiangqi.PieceType]string{
		xiangqi.General: "general", xiangqi.Advisor: "advisor", xiangqi.Elephant: "elephant",
		xiangqi.Horse: "horse", xiangqi.Rook: "rook", xiangqi.Cannon: "cannon", xiangqi.Pawn: "pawn",
	}
	return fmt.Sprintf("%s_%s", piece.Color.String(), names[piece.Type])
}
