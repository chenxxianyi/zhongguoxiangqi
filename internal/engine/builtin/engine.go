package builtin

import (
	"context"
	"errors"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/engine"
)

const (
	infinity          = 1_000_000
	mateScore         = 900_000
	mateThreshold     = mateScore - 1_000
	maxQuiescencePly  = 10
	maxSearchPly      = 128
	maxCheckExtension = 2
)

var (
	ErrNoLegalMove = errors.New("position has no legal move")
	pieceValues    = [8]int{0, 100_000, 200, 200, 420, 900, 450, 100}
)

type Engine struct{}

func New() *Engine { return &Engine{} }

func (*Engine) Name() string                 { return "builtin-alpha-beta" }
func (*Engine) Health(context.Context) error { return nil }
func (*Engine) Close() error                 { return nil }

type ttBound uint8

const (
	ttExact ttBound = iota + 1
	ttLower
	ttUpper
)

type ttEntry struct {
	key      uint64
	bestMove xiangqi.Move
	score    int
	depth    int
	bound    ttBound
}

type searchState struct {
	ctx      context.Context
	maxNodes uint64
	nodes    uint64
	stopped  bool
	table    []ttEntry
	killers  [maxSearchPly][2]xiangqi.Move
	history  [2][xiangqi.Ranks * xiangqi.Files][xiangqi.Ranks * xiangqi.Files]int
}

type rootLine struct {
	move  xiangqi.Move
	score int
	pv    []xiangqi.Move
}

type rankedMove struct {
	move  xiangqi.Move
	score int
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

	position := request.Position
	state := &searchState{
		ctx: ctx, maxNodes: request.MaxNodes,
		table: makeTranspositionTable(request.MaxNodes),
	}
	moves := orderedMoves(state, &position, 0, xiangqi.Move{})
	if len(moves) == 0 {
		return engine.AnalyzeResult{}, ErrNoLegalMove
	}

	var completed []rootLine
	completedDepth := 0
	for depth := 1; depth <= request.MaxDepth; depth++ {
		lines, ok := searchRoot(state, &position, moves, depth, request.MultiPV == 1)
		if !ok {
			break
		}
		completed = lines
		completedDepth = depth
		for index := range lines {
			moves[index] = lines[index].move
		}
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
		if position.SideToMove() == xiangqi.Black {
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

func searchRoot(state *searchState, position *xiangqi.Position, moves []xiangqi.Move, depth int, singlePV bool) ([]rootLine, bool) {
	lines := make([]rootLine, 0, len(moves))
	alpha := -infinity
	for index, move := range moves {
		if state.shouldStop() {
			return nil, false
		}
		next, _, err := position.Apply(move)
		if err != nil {
			continue
		}
		var score int
		var pv []xiangqi.Move
		var ok bool
		if !singlePV || index == 0 {
			score, pv, ok = negamax(state, &next, depth-1, -infinity, infinity, 1, 0)
			if !ok {
				return nil, false
			}
			score = -score
		} else {
			score, pv, ok = negamax(state, &next, depth-1, -alpha-1, -alpha, 1, 0)
			if !ok {
				return nil, false
			}
			score = -score
			if score > alpha {
				score, pv, ok = negamax(state, &next, depth-1, -infinity, -alpha, 1, 0)
				if !ok {
					return nil, false
				}
				score = -score
			}
		}
		if score > alpha {
			alpha = score
		}
		lines = append(lines, rootLine{move: move, score: score, pv: prependMove(move, pv)})
	}
	sortRootLines(lines)
	return lines, true
}

func negamax(state *searchState, position *xiangqi.Position, depth, alpha, beta, ply, checkExtensions int) (int, []xiangqi.Move, bool) {
	if depth <= 0 {
		return quiescence(state, position, alpha, beta, ply, 0)
	}
	if state.shouldStop() {
		return 0, nil, false
	}
	state.nodes++
	originalAlpha := alpha
	hash := position.Hash()
	ttMove := xiangqi.Move{}
	if entry, found := state.probe(hash); found {
		ttMove = entry.bestMove
		if entry.depth >= depth {
			score := scoreFromTT(entry.score, ply)
			switch entry.bound {
			case ttExact:
				return score, movePV(entry.bestMove), true
			case ttLower:
				if score > alpha {
					alpha = score
				}
			case ttUpper:
				if score < beta {
					beta = score
				}
			}
			if alpha >= beta {
				return score, movePV(entry.bestMove), true
			}
		}
	}

	moves := orderedMoves(state, position, ply, ttMove)
	if len(moves) == 0 {
		return -mateScore + ply, nil, true
	}
	best := -infinity
	bestMove := xiangqi.Move{}
	var bestPV []xiangqi.Move
	for moveIndex, move := range moves {
		next, captured, err := position.Apply(move)
		if err != nil {
			continue
		}
		nextDepth := depth - 1
		nextExtensions := checkExtensions
		if checkExtensions < maxCheckExtension && next.InCheck(next.SideToMove()) {
			nextDepth++
			nextExtensions++
		}
		searchAlpha, searchBeta := -beta, -alpha
		if moveIndex > 0 {
			searchAlpha, searchBeta = -alpha-1, -alpha
		}
		score, pv, ok := negamax(state, &next, nextDepth, searchAlpha, searchBeta, ply+1, nextExtensions)
		if !ok {
			return 0, nil, false
		}
		score = -score
		if moveIndex > 0 && score > alpha && score < beta {
			score, pv, ok = negamax(state, &next, nextDepth, -beta, -alpha, ply+1, nextExtensions)
			if !ok {
				return 0, nil, false
			}
			score = -score
		}
		if score > best {
			best = score
			bestMove = move
			bestPV = prependMove(move, pv)
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			if captured.Empty() {
				state.recordQuietCutoff(position.SideToMove(), move, depth, ply)
			}
			break
		}
	}
	bound := ttExact
	if best <= originalAlpha {
		bound = ttUpper
	} else if best >= beta {
		bound = ttLower
	}
	state.store(hash, depth, scoreToTT(best, ply), bound, bestMove)
	return best, bestPV, true
}

func quiescence(state *searchState, position *xiangqi.Position, alpha, beta, ply, qply int) (int, []xiangqi.Move, bool) {
	if state.shouldStop() {
		return 0, nil, false
	}
	state.nodes++
	if qply >= maxQuiescencePly || ply >= maxSearchPly-1 {
		return evaluateForTurn(position), nil, true
	}

	inCheck := position.InCheck(position.SideToMove())
	standPat := evaluateForTurn(position)
	if !inCheck {
		if standPat >= beta {
			return standPat, nil, true
		}
		if standPat > alpha {
			alpha = standPat
		}
	}

	moves := orderedMoves(state, position, ply, xiangqi.Move{})
	if len(moves) == 0 {
		return -mateScore + ply, nil, true
	}
	best := standPat
	var bestPV []xiangqi.Move
	searched := false
	for _, move := range moves {
		captured := position.PieceAt(move.To)
		if !inCheck && captured.Empty() {
			continue
		}
		searched = true
		next, _, err := position.Apply(move)
		if err != nil {
			continue
		}
		score, pv, ok := quiescence(state, &next, -beta, -alpha, ply+1, qply+1)
		if !ok {
			return 0, nil, false
		}
		score = -score
		if score > best || (inCheck && bestPV == nil) {
			best = score
			bestPV = prependMove(move, pv)
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break
		}
	}
	if inCheck && !searched {
		return -mateScore + ply, nil, true
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

func makeTranspositionTable(maxNodes uint64) []ttEntry {
	target := int(maxNodes / 4)
	if target < 4_096 {
		target = 4_096
	}
	if target > 262_144 {
		target = 262_144
	}
	size := 1
	for size < target {
		size <<= 1
	}
	return make([]ttEntry, size)
}

func (state *searchState) probe(hash uint64) (ttEntry, bool) {
	entry := state.table[int(hash&uint64(len(state.table)-1))]
	return entry, entry.bound != 0 && entry.key == hash
}

func (state *searchState) store(hash uint64, depth, score int, bound ttBound, bestMove xiangqi.Move) {
	index := int(hash & uint64(len(state.table)-1))
	existing := state.table[index]
	if existing.bound != 0 && existing.key == hash && existing.depth > depth {
		return
	}
	state.table[index] = ttEntry{
		key: hash, bestMove: bestMove, score: score, depth: depth, bound: bound,
	}
}

func scoreToTT(score, ply int) int {
	if score >= mateThreshold {
		return score + ply
	}
	if score <= -mateThreshold {
		return score - ply
	}
	return score
}

func scoreFromTT(score, ply int) int {
	if score >= mateThreshold {
		return score - ply
	}
	if score <= -mateThreshold {
		return score + ply
	}
	return score
}

func orderedMoves(state *searchState, position *xiangqi.Position, ply int, ttMove xiangqi.Move) []xiangqi.Move {
	moves := position.LegalMoves()
	ranked := make([]rankedMove, len(moves))
	for index, move := range moves {
		ranked[index] = rankedMove{move: move, score: state.moveScore(position, move, ply, ttMove)}
	}
	for index := 1; index < len(ranked); index++ {
		current := ranked[index]
		insert := index
		for insert > 0 && rankedMoveBefore(current, ranked[insert-1]) {
			ranked[insert] = ranked[insert-1]
			insert--
		}
		ranked[insert] = current
	}
	for index := range ranked {
		moves[index] = ranked[index].move
	}
	return moves
}

func (state *searchState) moveScore(position *xiangqi.Position, move xiangqi.Move, ply int, ttMove xiangqi.Move) int {
	if validMove(ttMove) && move == ttMove {
		return 2_000_000
	}
	attacker := position.PieceAt(move.From)
	captured := position.PieceAt(move.To)
	if !captured.Empty() {
		return 1_000_000 + pieceValues[captured.Type]*16 - pieceValues[attacker.Type]
	}
	if ply < maxSearchPly {
		if move == state.killers[ply][0] {
			return 800_000
		}
		if move == state.killers[ply][1] {
			return 700_000
		}
	}
	color := colorIndex(position.SideToMove())
	return state.history[color][squareIndex(move.From)][squareIndex(move.To)]
}

func (state *searchState) recordQuietCutoff(color xiangqi.Color, move xiangqi.Move, depth, ply int) {
	if ply < maxSearchPly && state.killers[ply][0] != move {
		state.killers[ply][1] = state.killers[ply][0]
		state.killers[ply][0] = move
	}
	value := &state.history[colorIndex(color)][squareIndex(move.From)][squareIndex(move.To)]
	*value += depth * depth
	if *value > 1_000_000 {
		*value = 1_000_000
	}
}

func rankedMoveBefore(left, right rankedMove) bool {
	if left.score != right.score {
		return left.score > right.score
	}
	return moveLess(left.move, right.move)
}

func sortRootLines(lines []rootLine) {
	for index := 1; index < len(lines); index++ {
		current := lines[index]
		insert := index
		for insert > 0 && rootLineBefore(current, lines[insert-1]) {
			lines[insert] = lines[insert-1]
			insert--
		}
		lines[insert] = current
	}
}

func rootLineBefore(left, right rootLine) bool {
	if left.score != right.score {
		return left.score > right.score
	}
	return moveLess(left.move, right.move)
}

func moveLess(left, right xiangqi.Move) bool {
	leftFrom, rightFrom := squareIndex(left.From), squareIndex(right.From)
	if leftFrom != rightFrom {
		return leftFrom < rightFrom
	}
	return squareIndex(left.To) < squareIndex(right.To)
}

func validMove(move xiangqi.Move) bool {
	return move.From.Valid() && move.To.Valid() && move.From != move.To
}

func movePV(move xiangqi.Move) []xiangqi.Move {
	if !validMove(move) {
		return nil
	}
	return []xiangqi.Move{move}
}

func prependMove(move xiangqi.Move, pv []xiangqi.Move) []xiangqi.Move {
	line := make([]xiangqi.Move, len(pv)+1)
	line[0] = move
	copy(line[1:], pv)
	return line
}

func squareIndex(square xiangqi.Square) int {
	return square.Rank*xiangqi.Files + square.File
}

func colorIndex(color xiangqi.Color) int {
	if color == xiangqi.Black {
		return 1
	}
	return 0
}

func evaluateForTurn(position *xiangqi.Position) int {
	score := evaluateRed(position)
	if position.SideToMove() == xiangqi.Black {
		return -score + 8
	}
	return score + 8
}

func evaluateRed(position *xiangqi.Position) int {
	score := 0
	for rank := 0; rank < xiangqi.Ranks; rank++ {
		for file := 0; file < xiangqi.Files; file++ {
			square := xiangqi.Square{File: file, Rank: rank}
			piece := position.PieceAt(square)
			if piece.Empty() {
				continue
			}
			value := pieceValues[piece.Type] + positionalBonus(position, piece, square)
			if piece.Color == xiangqi.Red {
				score += value
			} else {
				score -= value
			}
		}
	}
	return score
}

func positionalBonus(position *xiangqi.Position, piece xiangqi.Piece, square xiangqi.Square) int {
	fileCenter := 4 - abs(square.File-4)
	advance := square.Rank
	if piece.Color == xiangqi.Red {
		advance = 9 - square.Rank
	}
	switch piece.Type {
	case xiangqi.General:
		bonus := fileCenter * 3
		if advance == 0 {
			bonus += 12
		}
		return bonus
	case xiangqi.Advisor, xiangqi.Elephant:
		if advance <= 2 {
			return 8
		}
	case xiangqi.Horse:
		return fileCenter*7 + min(advance, 5)*3
	case xiangqi.Rook:
		return rayMobility(position, square) * 2
	case xiangqi.Cannon:
		return fileCenter*3 + rayMobility(position, square)
	case xiangqi.Pawn:
		bonus := advance*8 + fileCenter*2
		if advance >= 5 {
			bonus += 24
		}
		return bonus
	}
	return 0
}

func rayMobility(position *xiangqi.Position, square xiangqi.Square) int {
	directions := [...]xiangqi.Square{{File: 1}, {File: -1}, {Rank: 1}, {Rank: -1}}
	mobility := 0
	for _, direction := range directions {
		current := xiangqi.Square{File: square.File + direction.File, Rank: square.Rank + direction.Rank}
		for current.Valid() {
			mobility++
			if !position.PieceAt(current).Empty() {
				break
			}
			current.File += direction.File
			current.Rank += direction.Rank
		}
	}
	return mobility
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
