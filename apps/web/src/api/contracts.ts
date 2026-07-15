// ── 通用 ──

export interface ApiErrorBody {
  code: string
  message: string
  requestId?: string
  details?: unknown
}

// ── 对局 (Match) ──

export interface MoveRecord {
  ply: number
  move: string        // ICCS 格式，如 "a3a4"
  side: 'red' | 'black'
  actor: 'player' | 'ai'
  captured?: string
  fenBefore: string
  fenAfter: string
  hashAfter: string
  playedAt: string
  thinkTimeMs?: number
}

export type MatchStatus = 'active_player_turn' | 'active_ai_thinking' | 'finished' | 'aborted' | 'recoverable_error'

export type MatchOutcome = 'ongoing' | 'red_win' | 'black_win' | 'draw'
export type MatchTermination = '' | 'checkmate' | 'no_legal_moves' | 'resign' | 'draw_agreement' | 'threefold_repetition'

export interface MatchSnapshot {
  id: string
  version: number
  status: MatchStatus
  playerColor: 'red' | 'black'
  sideToMove: 'red' | 'black'
  difficulty: number
  engine: string
  allowUndo: boolean
  initialFen: string
  fen: string
  moves: MoveRecord[]
  outcome: MatchOutcome
  termination?: MatchTermination
  drawOffered: boolean
  createdAt: string
  updatedAt: string
}

export interface MatchEvent<T = unknown> {
  eventId: string
  matchId: string
  matchVersion: number
  type: string
  timestamp: string
  payload: T
}

export interface MatchEventPayloads {
  /* eslint-disable @typescript-eslint/no-empty-object-type */
  'match.snapshot': MatchSnapshot
  'match.move_accepted': MoveRecord
  'match.ai_thinking': { engine: string }
  'match.ai_move_applied': { move: MoveRecord; depth: number; nodes: number; stoppedReason: string }
  'match.finished': MatchSnapshot
  'match.undo_applied': MatchSnapshot
  'match.engine_degraded': { reason: string }
  'match.draw_declined': MatchSnapshot
  /* eslint-enable @typescript-eslint/no-empty-object-type */
}

// ── 难度配置 (Difficulty) ──

export interface DifficultyProfile {
  level: number
  name: string
  moveTimeMs: number
  maxDepth: number
  maxNodes: number
  multiPV: number
  description: string
}

// ── 棋谱记录 (Records) ──

export interface RecordMove {
  ply: number
  move: string
  side: string
  fenBefore: string
  fenAfter: string
  hashAfter: string
}

export interface GameRecord {
  id: string
  name: string
  format: string
  initialFen: string
  finalFen: string
  result?: string
  outcome: string
  termination?: string
  moveCount: number
  moves?: RecordMove[]
  tags?: string[]
  createdAt: string
}

export interface ImportError {
  ply?: number
  token?: string
  code: string
  message: string
}

export interface ImportBatch {
  id: string
  status: string
  name: string
  format: string
  totalGames: number
  importedGames: number
  duplicateGames: number
  failedGames: number
  recordIds: string[]
  errors: ImportError[]
  createdAt: string
  completedAt: string
}

// ── 学习 (Learning) ──

export interface LearningJob {
  id: string
  status: string
  name: string
  progress: number
  recordCount: number
  moveCount: number
  versionId?: string
  errorCode?: string
  message?: string
  createdAt: string
  completedAt?: string
}

export interface QualityReport {
  validRecords: number
  validMoves: number
  coveredPositions: number
  lowSampleEntries: number
}

export interface BookEntry {
  positionHash: string
  fen: string
  sideToMove: string
  move: string
  samples: number
  redWins: number
  blackWins: number
  draws: number
}

export interface LearningVersion {
  id: string
  name: string
  status: string
  algorithm: string
  quality: QualityReport
  entryCount: number
  entries?: BookEntry[]
  createdAt: string
  activatedAt?: string
}

// ── 复盘分析 (Analysis) ──

export interface AnalysisJob {
  id: string
  matchId: string
  status: string
  progress: number
  analyzedMoves: number
  totalMoves: number
  errorCode?: string
  message?: string
  createdAt: string
  completedAt?: string
}

export type MoveClassification = 'best' | 'excellent' | 'inaccuracy' | 'mistake' | 'blunder' | 'outside_top_candidates'

export interface MoveAnalysis {
  ply: number
  actualMove: string
  bestMove: string
  side: string
  classification: MoveClassification
  scoreLossCp?: number
  depth: number
  nodes: number
}

export interface AnalysisResult {
  matchId: string
  engine: string
  status: string
  analyzedMoves: number
  bestMoveRate: number
  moves: MoveAnalysis[]
  generatedAt: string
}
