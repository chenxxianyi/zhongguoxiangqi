package xiangqi

import (
	"errors"
	"fmt"
)

var (
	ErrIllegalMove = errors.New("illegal move")
	ErrWrongTurn   = errors.New("piece does not belong to side to move")
)

func (p *Position) IsLegal(move Move) bool {
	if !move.From.Valid() || !move.To.Valid() || move.From == move.To {
		return false
	}
	piece := p.PieceAt(move.From)
	if piece.Empty() || piece.Color != p.turn {
		return false
	}
	target := p.PieceAt(move.To)
	if !target.Empty() && (target.Color == piece.Color || target.Type == General) {
		return false
	}
	if !p.pseudoLegal(move, piece, target) {
		return false
	}
	next := p.applyUnchecked(move)
	return !next.InCheck(piece.Color)
}

func (p *Position) Apply(move Move) (Position, Piece, error) {
	piece := p.PieceAt(move.From)
	if piece.Empty() {
		return Position{}, Piece{}, fmt.Errorf("%w: origin is empty", ErrIllegalMove)
	}
	if piece.Color != p.turn {
		return Position{}, Piece{}, ErrWrongTurn
	}
	if !p.IsLegal(move) {
		return Position{}, Piece{}, ErrIllegalMove
	}
	captured := p.PieceAt(move.To)
	return p.applyUnchecked(move), captured, nil
}

func (p *Position) applyUnchecked(move Move) Position {
	next := *p
	next.board[move.To.Rank][move.To.File] = next.board[move.From.Rank][move.From.File]
	next.board[move.From.Rank][move.From.File] = Piece{}
	next.turn = p.turn.Opponent()
	next.ply++
	return next
}

func (p *Position) LegalMoves() []Move {
	moves := make([]Move, 0, 48)
	for rank := 0; rank < Ranks; rank++ {
		for file := 0; file < Files; file++ {
			from := Square{File: file, Rank: rank}
			piece := p.PieceAt(from)
			if piece.Empty() || piece.Color != p.turn {
				continue
			}
			for toRank := 0; toRank < Ranks; toRank++ {
				for toFile := 0; toFile < Files; toFile++ {
					move := Move{From: from, To: Square{File: toFile, Rank: toRank}}
					if p.IsLegal(move) {
						moves = append(moves, move)
					}
				}
			}
		}
	}
	return moves
}

func (p *Position) InCheck(color Color) bool {
	general, found := p.generalSquare(color)
	if !found {
		return true
	}
	attacker := color.Opponent()
	for rank := 0; rank < Ranks; rank++ {
		for file := 0; file < Files; file++ {
			from := Square{File: file, Rank: rank}
			piece := p.PieceAt(from)
			if piece.Empty() || piece.Color != attacker {
				continue
			}
			if p.pseudoLegal(Move{From: from, To: general}, piece, p.PieceAt(general)) {
				return true
			}
		}
	}
	return false
}

func (p *Position) Adjudicate() Adjudication {
	moves := p.LegalMoves()
	check := p.InCheck(p.turn)
	if len(moves) > 0 {
		return Adjudication{Outcome: OutcomeOngoing, InCheck: check, LegalMoves: len(moves)}
	}
	outcome := OutcomeRedWin
	if p.turn == Red {
		outcome = OutcomeBlackWin
	}
	reason := TerminationNoMoves
	if check {
		reason = TerminationCheckmate
	}
	return Adjudication{Outcome: outcome, Termination: reason, InCheck: check}
}

func (p *Position) pseudoLegal(move Move, piece, target Piece) bool {
	if !move.From.Valid() || !move.To.Valid() || move.From == move.To {
		return false
	}
	if !target.Empty() && target.Color == piece.Color {
		return false
	}
	df := move.To.File - move.From.File
	dr := move.To.Rank - move.From.Rank
	adf, adr := abs(df), abs(dr)

	switch piece.Type {
	case General:
		if target.Type == General && move.From.File == move.To.File {
			return p.blockersBetween(move.From, move.To) == 0
		}
		return adf+adr == 1 && inPalace(piece.Color, move.To)
	case Advisor:
		return adf == 1 && adr == 1 && inPalace(piece.Color, move.To)
	case Elephant:
		if adf != 2 || adr != 2 || !onOwnSide(piece.Color, move.To) {
			return false
		}
		eye := Square{File: (move.From.File + move.To.File) / 2, Rank: (move.From.Rank + move.To.Rank) / 2}
		return p.PieceAt(eye).Empty()
	case Horse:
		if !((adf == 2 && adr == 1) || (adf == 1 && adr == 2)) {
			return false
		}
		leg := move.From
		if adf == 2 {
			leg.File += sign(df)
		} else {
			leg.Rank += sign(dr)
		}
		return p.PieceAt(leg).Empty()
	case Rook:
		return (df == 0 || dr == 0) && p.blockersBetween(move.From, move.To) == 0
	case Cannon:
		if df != 0 && dr != 0 {
			return false
		}
		blockers := p.blockersBetween(move.From, move.To)
		if target.Empty() {
			return blockers == 0
		}
		return blockers == 1
	case Pawn:
		forward := -1
		crossed := move.From.Rank <= 4
		if piece.Color == Black {
			forward = 1
			crossed = move.From.Rank >= 5
		}
		if df == 0 && dr == forward {
			return true
		}
		return crossed && adr == 0 && adf == 1
	default:
		return false
	}
}

func (p *Position) blockersBetween(from, to Square) int {
	df, dr := sign(to.File-from.File), sign(to.Rank-from.Rank)
	if df != 0 && dr != 0 {
		return -1
	}
	count := 0
	for current := (Square{File: from.File + df, Rank: from.Rank + dr}); current != to; {
		if !p.PieceAt(current).Empty() {
			count++
		}
		current.File += df
		current.Rank += dr
	}
	return count
}

func (p *Position) generalSquare(color Color) (Square, bool) {
	for rank := 0; rank < Ranks; rank++ {
		for file := 0; file < Files; file++ {
			piece := p.board[rank][file]
			if piece.Type == General && piece.Color == color {
				return Square{File: file, Rank: rank}, true
			}
		}
	}
	return Square{}, false
}

func inPalace(color Color, square Square) bool {
	if square.File < 3 || square.File > 5 {
		return false
	}
	if color == Black {
		return square.Rank >= 0 && square.Rank <= 2
	}
	return square.Rank >= 7 && square.Rank <= 9
}

func onOwnSide(color Color, square Square) bool {
	if color == Black {
		return square.Rank <= 4
	}
	return square.Rank >= 5
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func sign(value int) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}
