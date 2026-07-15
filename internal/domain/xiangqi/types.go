package xiangqi

import (
	"errors"
	"fmt"
	"strings"
)

const (
	Files = 9
	Ranks = 10
)

const InitialFEN = "rheakaehr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RHEAKAEHR w"

type Color uint8

const (
	NoColor Color = iota
	Red
	Black
)

func (c Color) String() string {
	switch c {
	case Red:
		return "red"
	case Black:
		return "black"
	default:
		return "none"
	}
}

func (c Color) Opponent() Color {
	if c == Red {
		return Black
	}
	if c == Black {
		return Red
	}
	return NoColor
}

type PieceType uint8

const (
	NoPiece PieceType = iota
	General
	Advisor
	Elephant
	Horse
	Rook
	Cannon
	Pawn
)

type Piece struct {
	Color Color
	Type  PieceType
}

func (p Piece) Empty() bool { return p.Type == NoPiece }

type Square struct {
	File int `json:"file"`
	Rank int `json:"rank"`
}

func NewSquare(file, rank int) (Square, error) {
	sq := Square{File: file, Rank: rank}
	if !sq.Valid() {
		return Square{}, fmt.Errorf("square out of bounds: file=%d rank=%d", file, rank)
	}
	return sq, nil
}

func (s Square) Valid() bool {
	return s.File >= 0 && s.File < Files && s.Rank >= 0 && s.Rank < Ranks
}

// ICCS uses a0 at Red's left corner. Internally rank 0 is Black's back rank.
func (s Square) ICCS() string {
	if !s.Valid() {
		return ""
	}
	return string([]byte{byte('a' + s.File), byte('0' + (Ranks - 1 - s.Rank))})
}

func ParseSquare(raw string) (Square, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if len(raw) != 2 || raw[0] < 'a' || raw[0] > 'i' || raw[1] < '0' || raw[1] > '9' {
		return Square{}, fmt.Errorf("invalid ICCS square %q", raw)
	}
	return Square{File: int(raw[0] - 'a'), Rank: Ranks - 1 - int(raw[1]-'0')}, nil
}

type Move struct {
	From Square `json:"from"`
	To   Square `json:"to"`
}

func (m Move) ICCS() string {
	if !m.From.Valid() || !m.To.Valid() {
		return ""
	}
	return m.From.ICCS() + m.To.ICCS()
}

func ParseMove(raw string) (Move, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) != 4 {
		return Move{}, fmt.Errorf("invalid ICCS move %q", raw)
	}
	from, err := ParseSquare(raw[:2])
	if err != nil {
		return Move{}, err
	}
	to, err := ParseSquare(raw[2:])
	if err != nil {
		return Move{}, err
	}
	if from == to {
		return Move{}, errors.New("move origin and destination are identical")
	}
	return Move{From: from, To: to}, nil
}

type Outcome string

const (
	OutcomeOngoing  Outcome = "ongoing"
	OutcomeRedWin   Outcome = "red_win"
	OutcomeBlackWin Outcome = "black_win"
	OutcomeDraw     Outcome = "draw"
)

type Termination string

const (
	TerminationNone       Termination = ""
	TerminationCheckmate  Termination = "checkmate"
	TerminationNoMoves    Termination = "no_legal_moves"
	TerminationResign     Termination = "resign"
	TerminationAgreement  Termination = "draw_agreement"
	TerminationRepetition Termination = "threefold_repetition"
)

type Adjudication struct {
	Outcome     Outcome     `json:"outcome"`
	Termination Termination `json:"termination,omitempty"`
	InCheck     bool        `json:"inCheck"`
	LegalMoves  int         `json:"legalMoves"`
}
