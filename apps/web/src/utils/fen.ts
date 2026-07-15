import type { BoardPiece, Color, PieceName } from '@/types/xiangqi'

// ── FEN 字符 ↔ 棋子映射 ──

const FEN_CHAR_TO_PIECE: Record<string, { color: Color; name: PieceName }> = {
  K: { color: 'red', name: '帅' },
  k: { color: 'black', name: '将' },
  A: { color: 'red', name: '仕' },
  a: { color: 'black', name: '士' },
  B: { color: 'red', name: '相' },
  b: { color: 'black', name: '象' },
  E: { color: 'red', name: '相' },
  e: { color: 'black', name: '象' },
  H: { color: 'red', name: '马' },
  h: { color: 'black', name: '马' },
  N: { color: 'red', name: '马' },
  n: { color: 'black', name: '马' },
  R: { color: 'red', name: '车' },
  r: { color: 'black', name: '车' },
  C: { color: 'red', name: '炮' },
  c: { color: 'black', name: '炮' },
  P: { color: 'red', name: '兵' },
  p: { color: 'black', name: '卒' },
}

export const initialFEN = 'rheakaehr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RHEAKAEHR w'

/**
 * 解析 FEN 字符串为 BoardPiece[]。
 *
 * FEN 格式说明（与中国象棋的 ICCS 风格）：
 *   - 棋盘共 10 行 (rank 0-9)，9 列 (file 0-8)
 *   - Rank 0 = 黑方底线（棋盘上方），Rank 9 = 红方底线（棋盘下方）
 *   - 大写字母 = 红方，小写字母 = 黑方
 *   - 数字 = 连续空格数
 *   - "/" = 换行
 *   - 第二个字段 "w" = 红方行棋, "b" = 黑方行棋
 *
 * @param fen 标准的中国象棋 FEN 字符串
 * @returns BoardPiece[] 棋子数组
 */
export function parseFEN(fen: string): BoardPiece[] {
  const pieces: BoardPiece[] = []
  const boardPart = fen.split(' ')[0] ?? ''
  const ranks = boardPart.split('/')

  for (let rank = 0; rank < ranks.length; rank++) {
    const row = ranks[rank]!
    let file = 0
    for (const char of row) {
      if (char >= '1' && char <= '9') {
        file += parseInt(char, 10)
        continue
      }
      const mapping = FEN_CHAR_TO_PIECE[char]
      if (mapping) {
        const { color, name } = mapping
        pieces.push({
          id: `${color}-${name}-${file}-${rank}`,
          color,
          name,
          file,
          rank,
        })
      }
      file++
    }
  }

  return pieces
}

/**
 * 从 FEN 中提取行棋方。
 */
export function getSideToMove(fen: string): Color {
  const side = fen.split(' ')[1]
  return side === 'b' ? 'black' : 'red'
}

/**
 * 从 FEN 中提取局面哈希指纹（简化为 board 部分的 SHA-256 前 8 位）。
 * 用于判断局面是否变化。
 */
export function fenBoardHash(fen: string): string {
  const boardPart = fen.split(' ')[0] ?? ''
  // 简单的非加密哈希，仅用于 React 更新追踪
  let hash = 0
  for (let i = 0; i < boardPart.length; i++) {
    const chr = boardPart.charCodeAt(i)
    hash = ((hash << 5) - hash) + chr
    hash |= 0
  }
  return hash.toString(36)
}
