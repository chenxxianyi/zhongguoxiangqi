import type { BoardPiece, BoardSquare } from '@/types/xiangqi'

export function screenPosition(file: number, rank: number, flipped: boolean) {
  const displayFile = flipped ? 8 - file : file
  const displayRank = flipped ? 9 - rank : rank
  return {
    left: `${4.5 + displayFile * 11.375}%`,
    top: `${4.5 + displayRank * 10.11}%`,
  }
}

export function candidateMoves(piece: BoardPiece): BoardSquare[] {
  const moves: BoardSquare[] = []
  const push = (file: number, rank: number) => {
    if (file >= 0 && file <= 8 && rank >= 0 && rank <= 9) moves.push({ file, rank })
  }
  const direction = piece.color === 'red' ? -1 : 1

  if (piece.name === '马') {
    ;[[-2, -1], [-2, 1], [-1, -2], [-1, 2]].forEach(([df = 0, dr = 0]) => push(piece.file + df, piece.rank + dr))
  } else if (piece.name === '车' || piece.name === '炮') {
    push(piece.file, piece.rank + direction)
    push(piece.file, piece.rank + direction * 2)
    push(piece.file - 1, piece.rank)
    push(piece.file + 1, piece.rank)
  } else {
    push(piece.file, piece.rank + direction)
    push(piece.file - 1, piece.rank)
    push(piece.file + 1, piece.rank)
  }
  return moves.slice(0, 4)
}
