package builtin

import (
	"context"
	"errors"
	"sort"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/engine"
)

const (
	infinity  = 1_000_000
	mateScore = 900_000
)

var ErrNoLegalMove = errors.New("position has no legal move")

type Engine struct{}

func New() *Engine { return &Engine{} }

func (*Engine) Name() string                 { return "builtin-alpha-beta" }
func (*Engine) Health(context.Context) error { return nil }
func (*Engine) Close() error                 { return nil }

type searchState struct {
	ctx      context.Context
	maxNodes uint64
	nodes    uint64
	stopped  bool
}

type rootLine struct {
	move  xiangqi.Move
	score int
	pv    []xiangqi.Move
}

func (*Engine) Analyze(ctx context.Context, request engine.AnalyzeRequest) (engine.AnalyzeResult, error) {
	started := time.Now()
	if request.MaxDepth <= 0 {
		request.MaxDepth = 3
	}
	if request.MaxNodes == 0 {
		request.MaxNodes = 50_000
	}
	if request.MultiPV <= 0 {
		request.MultiPV = 1
	}
	if request.MoveTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.MoveTime)
		defer cancel()
	}
	moves := orderedMoves(request.Position)
	if len(moves) == 0 {
		return engine.AnalyzeResult{}, ErrNoLegalMove
	}

	state := &searchState{ctx: ctx, maxNodes: request.MaxNodes}
	var completed []rootLine
	completedDepth := 0
	for depth := 1; depth <= request.MaxDepth; depth++ {
		lines, ok := searchRoot(state, request.Position, moves, depth)
		if !ok {
			break
		}
		completed = lines
		completedDepth = depth
		if len(lines) > 0 && abs(lines[0].score) >= mateScore-100 {
			break
		}
	}
	if len(completed) == 0 {
		// A tiny budget may stop before depth one completes. Return a legal move,
		// but explicitly report the incomplete stop instead of inventing a score.
		completed = []rootLine{{move: moves[0]}}
	}
	limit := min(request.MultiPV, len(completed))
	candidates := make([]engine.Candidate, 0, limit)
	for _, line := range completed[:limit] {
		score := line.score
		if request.Position.SideToMove() == xiangqi.Black {
			score = -score
		}
		candidates = append(candidates, engine.Candidate{
			Move: line.move, MoveICCS: line.move.ICCS(), ScoreCP: score, PV: line.pv,
		})
	}
	reason := "depth_limit"
	if state.stopped {
		if ctx.Err() != nil {
			reason = "time_or_context"
		} else {
			reason = "node_limit"
		}
	}
	return engine.AnalyzeResult{
		BestMove: completed[0].move, BestMoveICCS: completed[0].move.ICCS(),
		Candidates: candidates, Depth: completedDepth, Nodes: state.nodes,
		Duration: time.Since(started), StoppedReason: reason,
	}, nil
}

func searchRoot(state *searchState, position xiangqi.Position, moves []xiangqi.Move, depth int) ([]rootLine, bool) {
	lines := make([]rootLine, 0, len(moves))
	for _, move := range moves {
		if state.shouldStop() {
			return nil, false
		}
		next, _, err := position.Apply(move)
		if err != nil {
			continue
		}
		score, pv, ok := negamax(state, next, depth-1, -infinity, infinity, 1)
		if !ok {
			return nil, false
		}
		lines = append(lines, rootLine{move: move, score: -score, pv: append([]xiangqi.Move{move}, pv...)})
	}
	sort.SliceStable(lines, func(i, j int) bool {
		if lines[i].score == lines[j].score {
			return lines[i].move.ICCS() < lines[j].move.ICCS()
		}
		return lines[i].score > lines[j].score
	})
	return lines, true
}

func negamax(state *searchState, position xiangqi.Position, depth, alpha, beta, ply int) (int, []xiangqi.Move, bool) {
	if state.shouldStop() {
		return 0, nil, false
	}
	state.nodes++
	moves := orderedMoves(position)
	if len(moves) == 0 {
		return -mateScore + ply, nil, true
	}
	if depth == 0 {
		return evaluateForTurn(position), nil, true
	}
	best := -infinity
	var bestPV []xiangqi.Move
	for _, move := range moves {
		next, _, err := position.Apply(move)
		if err != nil {
			continue
		}
		score, pv, ok := negamax(state, next, depth-1, -beta, -alpha, ply+1)
		if !ok {
			return 0, nil, false
		}
		score = -score
		if score > best {
			best = score
			bestPV = append([]xiangqi.Move{move}, pv...)
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break
		}
	}
	return best, bestPV, true
}

func (state *searchState) shouldStop() bool {
	if state.stopped {
		return true
	}
	if state.nodes >= state.maxNodes {
		state.stopped = true
		return true
	}
	select {
	case <-state.ctx.Done():
		state.stopped = true
		return true
	default:
		return false
	}
}

func orderedMoves(position xiangqi.Position) []xiangqi.Move {
	moves := position.LegalMoves()
	sort.SliceStable(moves, func(i, j int) bool {
		leftCapture := !position.PieceAt(moves[i].To).Empty()
		rightCapture := !position.PieceAt(moves[j].To).Empty()
		if leftCapture != rightCapture {
			return leftCapture
		}
		return moves[i].ICCS() < moves[j].ICCS()
	})
	return moves
}

func evaluateForTurn(position xiangqi.Position) int {
	score := evaluateRed(position)
	if position.SideToMove() == xiangqi.Black {
		return -score
	}
	return score
}

func evaluateRed(position xiangqi.Position) int {
	values := map[xiangqi.PieceType]int{
		xiangqi.General:  100_000,
		xiangqi.Rook:     900,
		xiangqi.Cannon:   450,
		xiangqi.Horse:    420,
		xiangqi.Elephant: 200,
		xiangqi.Advisor:  200,
		xiangqi.Pawn:     100,
	}
	score := 0
	for rank := 0; rank < xiangqi.Ranks; rank++ {
		for file := 0; file < xiangqi.Files; file++ {
			piece := position.PieceAt(xiangqi.Square{File: file, Rank: rank})
			if piece.Empty() {
				continue
			}
			value := values[piece.Type]
			if piece.Type == xiangqi.Pawn {
				if piece.Color == xiangqi.Red {
					value += (9 - rank) * 6
				} else {
					value += rank * 6
				}
			}
			if piece.Color == xiangqi.Red {
				score += value
			} else {
				score -= value
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
