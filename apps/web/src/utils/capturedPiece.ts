const CAPTURED_PIECE_LABELS: Record<string, string> = {
  red_general: '帅',
  black_general: '将',
  red_advisor: '仕',
  black_advisor: '士',
  red_elephant: '相',
  black_elephant: '象',
  red_horse: '马',
  black_horse: '马',
  red_rook: '车',
  black_rook: '车',
  red_cannon: '炮',
  black_cannon: '炮',
  red_pawn: '兵',
  black_pawn: '卒',
}

const CHINESE_PIECE_LABEL = /^[帅将仕士相象马车炮兵卒]$/

export function getCapturedPieceLabel(captured: string | undefined): string {
  if (!captured) return ''
  if (CHINESE_PIECE_LABEL.test(captured)) return captured
  return CAPTURED_PIECE_LABELS[captured.toLowerCase()] ?? captured
}
