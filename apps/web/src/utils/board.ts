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

export function fromICCSSquare(square: string): BoardSquare | null {
  if (!/^[a-i][0-9]$/i.test(square)) return null
  return {
    file: square.toLowerCase().charCodeAt(0) - 97,
    rank: 9 - Number(square[1]),
  }
}

/**
 * 将两个坐标转为 ICCS 着法（如 "a3a4"）。
 */
export function toICCSCode(fromFile: number, fromRank: number, toFile: number, toRank: number): string {
  return toICCSSquare(fromFile, fromRank) + toICCSSquare(toFile, toRank)
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
