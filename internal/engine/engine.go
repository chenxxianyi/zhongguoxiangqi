package engine

import (
	"context"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
)

type AnalyzeRequest struct {
	Position xiangqi.Position
	MaxDepth int
	MaxNodes uint64
	MoveTime time.Duration
	MultiPV  int
}

type Candidate struct {
	Move     xiangqi.Move   `json:"move"`
	MoveICCS string         `json:"moveIccs"`
	ScoreCP  int            `json:"scoreCp"`
	PV       []xiangqi.Move `json:"pv,omitempty"`
}

type AnalyzeResult struct {
	BestMove      xiangqi.Move  `json:"bestMove"`
	BestMoveICCS  string        `json:"bestMoveIccs"`
	Candidates    []Candidate   `json:"candidates"`
	Depth         int           `json:"depth"`
	Nodes         uint64        `json:"nodes"`
	Duration      time.Duration `json:"duration"`
	StoppedReason string        `json:"stoppedReason"`
}

type Engine interface {
	Name() string
	Analyze(context.Context, AnalyzeRequest) (AnalyzeResult, error)
	Health(context.Context) error
	Close() error
}
