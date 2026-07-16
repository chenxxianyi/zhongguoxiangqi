/**
 * ICCS 坐标转中文列号（红方视角从右往左 1-9，黑方视角从左往右 1-9）。
 * 在 ICCS 中 a=0, b=1, ..., i=8。
 */
function iccsFileToChineseCol(file: number, viewerSide: 'red' | 'black'): number {
  return viewerSide === 'red' ? 9 - file : file + 1
}

/**
 * 棋子中文名映射（用于简写 FEN 字符 → 中文名）。
 */
const PIECE_DISPLAY: Record<string, string> = {
  K: '帅', k: '将',
  A: '仕', a: '士',
  B: '相', b: '象',
  E: '相', e: '象',
  H: '马', h: '马',
  N: '马', n: '马',
  R: '车', r: '车',
  C: '炮', c: '炮',
  P: '兵', p: '卒',
}

/**
 * 简单转换：将 ICCS 着法（如 "b0c0"）转为可读格式。
 *
 * 基本格式：`车二平五` 风格（红方视角）。
 * 如果无法精确转换，回退为 "file→file" 格式。
 *
 * @param iccsMove 4 字符 ICCS 着法，如 "a3a4"
 * @param fen 当前 FEN（用于查找棋子名）
 * @param viewerSide 观看视角
 * @returns 可读的中文着法字符串
 */
export function iccsToDisplay(iccsMove: string, fen?: string, viewerSide: 'red' | 'black' = 'red'): string {
  if (iccsMove.length !== 4) return iccsMove

  const fromFile = iccsMove.charCodeAt(0) - 97
  const fromRank = 9 - parseInt(iccsMove[1]!, 10)
  const toFile = iccsMove.charCodeAt(2) - 97
  const toRank = 9 - parseInt(iccsMove[3]!, 10)

  // 从 FEN 中查找该位置的棋子名
  let pieceName = ''
  if (fen) {
    const ranks = fen.split(' ')[0]?.split('/') ?? []
    const row = ranks[fromRank]
    if (row) {
      let col = 0
      for (const char of row) {
        if (char >= '1' && char <= '9') {
          col += parseInt(char, 10)
          continue
        }
        if (col === fromFile) {
          pieceName = PIECE_DISPLAY[char] ?? '?'
          break
        }
        col++
      }
    }
  }

  const fromCol = iccsFileToChineseCol(fromFile, viewerSide)
  const toCol = iccsFileToChineseCol(toFile, viewerSide)
  const toRow = viewerSide === 'red' ? 10 - toRank : toRank + 1

  // 根据棋子的中文名和移动类型构建着法描述
  const prefix = pieceName || `(${String.fromCharCode(97 + fromFile)}${fromRank})`

  // 简单分类：平移（同行）、前进/后退
  if (fromRank === toRank) {
    return `${prefix}${fromCol}平${toCol}`
  } else if (viewerSide === 'red' ? toRank < fromRank : toRank > fromRank) {
    return `${prefix}${fromCol}进${toCol === fromCol ? toRow : toCol}`
  } else {
    return `${prefix}${fromCol}退${toCol === fromCol ? toRow : toCol}`
  }
}

/**
 * 将 ICCS 着法转为简洁格式，如 "a3→a4"。
 */
export function iccsToBrief(iccsMove: string): string {
  if (iccsMove.length !== 4) return iccsMove
  return `${iccsMove.slice(0, 2)}→${iccsMove.slice(2)}`
}

/**
 * 生成简化中文着法表（红方视角）。
 * 用于对局面板中的着法列表显示。
 */
export function formatMoveList(
  moves: Array<{ ply: number; move: string; side: string; actor: string }>,
  fen?: string,
): Array<[string, string, string]> {
  const result: Array<[string, string, string]> = []
  let currentFen = fen

  for (let i = 0; i < moves.length; i += 2) {
    const num = `${Math.floor(i / 2) + 1}.`
    const redMove = moves[i] ? iccsToDisplay(moves[i]!.move, currentFen, 'red') : ''
    const blackMove = moves[i + 1] ? iccsToDisplay(moves[i + 1]!.move, currentFen, 'black') : ''

    result.push([num, redMove, blackMove])

    // 更新 FEN（如果提供了棋盘信息）
    if (currentFen && moves[i]?.fenAfter) {
      currentFen = moves[i]!.fenAfter
    }
    if (currentFen && moves[i + 1]?.fenAfter) {
      currentFen = moves[i + 1]!.fenAfter
    }
  }

  return result
}
