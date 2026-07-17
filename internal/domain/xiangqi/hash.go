package xiangqi

var (
	zobristPieces [2][7][Ranks * Files]uint64
	zobristTurn   uint64
)

func init() {
	state := uint64(0x7869616e67716921)
	next := func() uint64 {
		state += 0x9e3779b97f4a7c15
		value := state
		value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
		value = (value ^ (value >> 27)) * 0x94d049bb133111eb
		return value ^ (value >> 31)
	}
	for color := range zobristPieces {
		for kind := range zobristPieces[color] {
			for square := range zobristPieces[color][kind] {
				zobristPieces[color][kind][square] = next()
			}
		}
	}
	zobristTurn = next()
}

func (p *Position) Hash() uint64 {
	var hash uint64
	for rank := 0; rank < Ranks; rank++ {
		for file := 0; file < Files; file++ {
			piece := p.board[rank][file]
			if piece.Empty() {
				continue
			}
			colorIndex := 0
			if piece.Color == Black {
				colorIndex = 1
			}
			hash ^= zobristPieces[colorIndex][piece.Type-1][rank*Files+file]
		}
	}
	if p.turn == Black {
		hash ^= zobristTurn
	}
	return hash
}
