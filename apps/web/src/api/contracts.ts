export interface ApiErrorBody {
  code: string
  message: string
  requestId?: string
  details?: unknown
}

export interface MatchEvent<T = unknown> {
  eventId: string
  matchId: string
  matchVersion: number
  type: string
  timestamp: string
  payload: T
}

export interface MatchSnapshot {
  id: string
  version: number
  status: string
  sideToMove: 'red' | 'black'
  fen: string
}
