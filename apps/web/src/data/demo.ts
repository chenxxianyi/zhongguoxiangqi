import type { BoardPiece, DifficultyProfile, HistoryMatch } from '@/types/xiangqi'

export const difficultyProfiles: DifficultyProfile[] = [
  { group: '入门', name: '认识棋局', time: '0.2–0.5 秒', multiPv: '3 路', randomness: '较高', description: '适合熟悉规则，会在合理着法中保留明显容错。' },
  { group: '入门', name: '轻松起步', time: '0.3–0.7 秒', multiPv: '3 路', randomness: '较高', description: '偏向清晰直观的走法，避免无意义送子。' },
  { group: '休闲', name: '从容对弈', time: '0.5–1 秒', multiPv: '4 路', randomness: '中等', description: '具备基础战术意识，候选着变化较丰富。' },
  { group: '休闲', name: '有来有回', time: '0.8–1.5 秒', multiPv: '4 路', randomness: '中等', description: '会把握简单战机，同时保留适度随机性。' },
  { group: '进阶', name: '谨慎谋划', time: '1–2 秒', multiPv: '4 路', randomness: '较低', description: '开始关注中局结构和子力协调。' },
  { group: '进阶', name: '沉稳应战', time: '1.5–3 秒', multiPv: '4 路', randomness: '较低', description: '会进行更深入的局面判断，在多个合理候选着中保持少量变化。' },
  { group: '高手', name: '精确计算', time: '2–4 秒', multiPv: '5 路', randomness: '很低', description: '缩小候选评分带，重视战术与局面转换。' },
  { group: '高手', name: '深度布局', time: '3–6 秒', multiPv: '5 路', randomness: '很低', description: '更深入地搜索复杂变化，较少主动放弃优势。' },
  { group: '大师', name: '强力挑战', time: '5–8 秒', multiPv: '6 路', randomness: '极低', description: '使用更高搜索预算，优先选择高质量候选着。' },
  { group: '大师', name: '极致棋力', time: '8–12 秒', multiPv: '8 路', randomness: '极低', description: '接近当前配置的最高搜索资源，不标注未经校准的 Elo。' },
]

const piece = (id: string, color: BoardPiece['color'], name: BoardPiece['name'], file: number, rank: number, last = false): BoardPiece => ({ id, color, name, file, rank, last })

export const initialPieces: BoardPiece[] = [
  piece('b-r1', 'black', '车', 0, 0), piece('b-h1', 'black', '马', 1, 0), piece('b-e1', 'black', '象', 2, 0), piece('b-a1', 'black', '士', 3, 0), piece('b-k', 'black', '将', 4, 0), piece('b-a2', 'black', '士', 5, 0), piece('b-e2', 'black', '象', 6, 0), piece('b-h2', 'black', '马', 7, 0), piece('b-r2', 'black', '车', 8, 0),
  piece('b-c1', 'black', '炮', 1, 2), piece('b-c2', 'black', '炮', 7, 2), piece('b-p1', 'black', '卒', 0, 3), piece('b-p2', 'black', '卒', 2, 3), piece('b-p3', 'black', '卒', 4, 3), piece('b-p4', 'black', '卒', 6, 3), piece('b-p5', 'black', '卒', 8, 3),
  piece('r-p1', 'red', '兵', 0, 6), piece('r-p2', 'red', '兵', 2, 6), piece('r-p3', 'red', '兵', 4, 6), piece('r-p4', 'red', '兵', 6, 6), piece('r-p5', 'red', '兵', 8, 6, true), piece('r-c1', 'red', '炮', 1, 7), piece('r-c2', 'red', '炮', 7, 7),
  piece('r-r1', 'red', '车', 0, 9), piece('r-h1', 'red', '马', 1, 9), piece('r-e1', 'red', '相', 2, 9), piece('r-a1', 'red', '仕', 3, 9), piece('r-k', 'red', '帅', 4, 9), piece('r-a2', 'red', '仕', 5, 9), piece('r-e2', 'red', '相', 6, 9), piece('r-h2', 'red', '马', 7, 9), piece('r-r2', 'red', '车', 8, 9),
]

export const historyMatches: HistoryMatch[] = [
  { resultClass: 'win', result: '胜', opponent: 'AI · 休闲 4', opening: '中炮对屏风马 · 执红', mode: '标准引擎', length: '42 回合', bookHit: '38%', date: '今日 09:42' },
  { resultClass: 'loss', result: '负', opponent: 'AI · 进阶 6', opening: '仙人指路 · 执黑', mode: '棋风模仿', length: '56 回合', bookHit: '26%', date: '昨日 21:18' },
  { resultClass: 'draw', result: '和', opponent: 'AI · 进阶 5', opening: '飞相局 · 执红', mode: '棋谱优先', length: '68 回合', bookHit: '44%', date: '7 月 13 日' },
  { resultClass: 'win', result: '胜', opponent: 'AI · 入门 2', opening: '顺炮直车 · 执黑', mode: '标准引擎', length: '31 回合', bookHit: '0%', date: '7 月 12 日' },
  { resultClass: 'win', result: '胜', opponent: 'AI · 高手 7', opening: '中炮急进中兵 · 执红', mode: '棋谱优先', length: '49 回合', bookHit: '33%', date: '7 月 10 日' },
]
