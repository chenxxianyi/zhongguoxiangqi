package xiangqi

import (
	"fmt"
	"strconv"
	"strings"
)

type Position struct {
	board [Ranks][Files]Piece
	turn  Color
	ply   int
}

func InitialPosition() Position {
	position, err := ParseFEN(InitialFEN)
	if err != nil {
		panic(err)
	}
	return position
}

func ParseFEN(raw string) (Position, error) {
	if len(raw) == 0 || len(raw) > 256 {
		return Position{}, fmt.Errorf("FEN length is invalid")
	}
	fields := strings.Fields(raw)
	if len(fields) < 2 || len(fields) > 6 {
		return Position{}, fmt.Errorf("FEN must have 2 to 6 fields")
	}
	rows := strings.Split(fields[0], "/")
	if len(rows) != Ranks {
		return Position{}, fmt.Errorf("FEN board must contain %d ranks", Ranks)
	}
	var position Position
	generals := map[Color]int{}
	for rank, row := range rows {
		file := 0
		for _, char := range row {
			if char >= '1' && char <= '9' {
				file += int(char - '0')
				continue
			}
			piece, ok := pieceFromFEN(char)
			if !ok {
				return Position{}, fmt.Errorf("invalid FEN piece %q", char)
			}
			if file >= Files {
				return Position{}, fmt.Errorf("too many files at rank %d", rank)
			}
			position.board[rank][file] = piece
			if piece.Type == General {
				generals[piece.Color]++
			}
			file++
		}
		if file != Files {
			return Position{}, fmt.Errorf("rank %d has %d files, expected %d", rank, file, Files)
		}
	}
	switch fields[1] {
	case "w", "r":
		position.turn = Red
	case "b":
		position.turn = Black
	default:
		return Position{}, fmt.Errorf("invalid FEN side to move %q", fields[1])
	}
	if generals[Red] != 1 || generals[Black] != 1 {
		return Position{}, fmt.Errorf("FEN must contain exactly one general for each side")
	}
	if len(fields) >= 6 {
		fullmove, err := strconv.Atoi(fields[5])
		if err == nil && fullmove > 0 {
			position.ply = (fullmove - 1) * 2
			if position.turn == Black {
				position.ply++
			}
		}
	}
	return position, nil
}

func (p Position) FEN() string {
	var rows [Ranks]string
	for rank := 0; rank < Ranks; rank++ {
		var b strings.Builder
		empty := 0
		for file := 0; file < Files; file++ {
			piece := p.board[rank][file]
			if piece.Empty() {
				empty++
				continue
			}
			if empty > 0 {
				b.WriteByte(byte('0' + empty))
				empty = 0
			}
			b.WriteRune(pieceFEN(piece))
		}
		if empty > 0 {
			b.WriteByte(byte('0' + empty))
		}
		rows[rank] = b.String()
	}
	side := "w"
	if p.turn == Black {
		side = "b"
	}
	return strings.Join(rows[:], "/") + " " + side
}

func (p Position) SideToMove() Color { return p.turn }
func (p Position) Ply() int          { return p.ply }

func (p Position) PieceAt(square Square) Piece {
	if !square.Valid() {
		return Piece{}
	}
	return p.board[square.Rank][square.File]
}

func pieceFromFEN(char rune) (Piece, bool) {
	color := Black
	if char >= 'A' && char <= 'Z' {
		color = Red
		char += 'a' - 'A'
	}
	var kind PieceType
	switch char {
	case 'k':
		kind = General
	case 'a':
		kind = Advisor
	case 'e', 'b':
		kind = Elephant
	case 'h', 'n':
		kind = Horse
	case 'r':
		kind = Rook
	case 'c':
		kind = Cannon
	case 'p':
		kind = Pawn
	default:
		return Piece{}, false
	}
	return Piece{Color: color, Type: kind}, true
}

func pieceFEN(piece Piece) rune {
	var char rune
	switch piece.Type {
	case General:
		char = 'k'
	case Advisor:
		char = 'a'
	case Elephant:
		char = 'e'
	case Horse:
		char = 'h'
	case Rook:
		char = 'r'
	case Cannon:
		char = 'c'
	case Pawn:
		char = 'p'
	default:
		return '?'
	}
	if piece.Color == Red {
		char -= 'a' - 'A'
	}
	return char
}
