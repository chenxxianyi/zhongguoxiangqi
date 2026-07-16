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
