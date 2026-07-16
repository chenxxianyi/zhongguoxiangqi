import type { BoardPiece, BoardSquare } from '@/types/xiangqi'
import { parseFEN } from './fen'

export function screenPosition(file: number, rank: number, flipped: boolean) {
  const displayFile = flipped ? 8 - file : file
  const displayRank = flipped ? 9 - rank : rank
  return {
    left: `${4.5 + displayFile * 11.375}%`,
    top: `${4.5 + displayRank * 10.11}%`,
  }
}

/**
 * 将 file/rank 转为 ICCS 坐标字符串（如 a0, i9）。
 */
export function toICCSSquare(file: number, rank: number): string {
  return String.fromCharCode(97 + file) + (9 - rank)
}

/**
 * 将两个坐标转为 ICCS 着法（如 "a3a4"）。
 */
export function toICCSCode(fromFile: number, fromRank: number, toFile: number, toRank: number): string {
  return toICCSSquare(fromFile, fromRank) + toICCSSquare(toFile, toRank)
}

/**
 * 简化的候选着法提示 —— 用于 UI 交互参考。
 * 实际所有着法合法性均由服务端验证，此处仅为落点预览。
 */
export function candidateMoves(piece: BoardPiece): BoardSquare[] {
  const moves: BoardSquare[] = []
  const push = (file: number, rank: number) => {
    if (file >= 0 && file <= 8 && rank >= 0 && rank <= 9) moves.push({ file, rank })
  }
  const direction = piece.color === 'red' ? -1 : 1

  // 马（蹩脚逻辑在此简化，仅演示）
  if (piece.name === '马') {
    push(piece.file - 2, piece.rank - 1)
    push(piece.file - 2, piece.rank + 1)
    push(piece.file - 1, piece.rank - 2)
    push(piece.file - 1, piece.rank + 2)
    push(piece.file + 1, piece.rank - 2)
    push(piece.file + 1, piece.rank + 2)
    push(piece.file + 2, piece.rank - 1)
    push(piece.file + 2, piece.rank + 1)
  } else if (piece.name === '车' || piece.name === '炮') {
    // 车、炮：直线走法（简化，不阻塞）
    for (let i = 1; i <= 9; i++) {
      push(piece.file, piece.rank + i)
      push(piece.file, piece.rank - i)
      push(piece.file + i, piece.rank)
      push(piece.file - i, piece.rank)
    }
  } else if (piece.name === '将' || piece.name === '帅') {
    push(piece.file, piece.rank + direction)
    push(piece.file, piece.rank - direction)
    push(piece.file - 1, piece.rank)
    push(piece.file + 1, piece.rank)
  } else if (piece.name === '仕' || piece.name === '士') {
    push(piece.file - 1, piece.rank + direction)
    push(piece.file + 1, piece.rank + direction)
  } else if (piece.name === '相' || piece.name === '象') {
    push(piece.file - 2, piece.rank + direction * 2)
    push(piece.file + 2, piece.rank + direction * 2)
  } else if (piece.name === '兵' || piece.name === '卒') {
    // 兵/卒：未过河只能前进，过河可左右
    const crossedRiver = piece.color === 'red' ? piece.rank <= 4 : piece.rank >= 5
    push(piece.file, piece.rank + direction)
    if (crossedRiver) {
      push(piece.file - 1, piece.rank)
      push(piece.file + 1, piece.rank)
    }
  } else {
    // 其他棋子（炮的斜线等 - 简化）
    push(piece.file, piece.rank + direction)
    push(piece.file - 1, piece.rank)
    push(piece.file + 1, piece.rank)
  }
  return moves
}

/**
 * 从 FEN 构建 BoardPiece[]，并覆写 last 标记。
 */
export function getPiecesFromFEN(
  fen: string,
  lastMoveIccs?: string,
): BoardPiece[] {
  const pieces = parseFEN(fen)

  if (lastMoveIccs && lastMoveIccs.length === 4) {
    const toFile = lastMoveIccs.charCodeAt(2) - 97
    const toRank = 9 - parseInt(lastMoveIccs[3]!, 10)
    for (const piece of pieces) {
      if (piece.file === toFile && piece.rank === toRank) {
        piece.last = true
        break
      }
    }
  }

  return pieces
}
