export type Color = 'red' | 'black'
export type PieceName = '车' | '马' | '相' | '象' | '仕' | '士' | '帅' | '将' | '炮' | '兵' | '卒'

export interface BoardPiece {
  id: string
  color: Color
  name: PieceName
  file: number
  rank: number
  last?: boolean
}

export interface BoardSquare {
  file: number
  rank: number
}

export type SideChoice = 'red' | 'black' | 'random'
export type AiMode = 'standard' | 'library' | 'style'

export interface DifficultyProfile {
  group: '入门' | '休闲' | '进阶' | '高手' | '大师'
  name: string
  time: string
  multiPv: string
  randomness: string
  description: string
}

export interface HistoryMatch {
  resultClass: 'win' | 'loss' | 'draw'
  result: '胜' | '负' | '和'
  opponent: string
  opening: string
  mode: string
  length: string
  bookHit: string
  date: string
}
