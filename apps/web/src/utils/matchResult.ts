import type {
  MatchOutcome,
  MatchSnapshot,
  MatchStatus,
  MatchTermination,
} from '@/api/contracts'

export type PlayerMatchResult = 'win' | 'loss' | 'draw' | 'ongoing' | 'aborted'

export interface MatchResultStats {
  total: number
  finished: number
  wins: number
  draws: number
  losses: number
  winRate: number
}

export function isActiveMatch(status: MatchStatus): boolean {
  return status !== 'finished' && status !== 'aborted'
}

export function getPlayerResult(
  outcome: MatchOutcome,
  playerColor: MatchSnapshot['playerColor'],
): PlayerMatchResult {
  if (outcome === 'ongoing') return 'ongoing'
  if (outcome === 'draw') return 'draw'

  const playerWon = (
    outcome === 'red_win' && playerColor === 'red'
  ) || (
    outcome === 'black_win' && playerColor === 'black'
  )
  return playerWon ? 'win' : 'loss'
}

export function getMatchResult(match: MatchSnapshot): PlayerMatchResult {
  if (isActiveMatch(match.status)) return 'ongoing'
  if (match.status === 'aborted') return 'aborted'
  return getPlayerResult(match.outcome, match.playerColor)
}

export function getPlayerResultLabel(result: PlayerMatchResult): string {
  return {
    win: '胜',
    loss: '负',
    draw: '和',
    ongoing: '续',
    aborted: '止',
  }[result]
}

export function getPlayerResultClass(result: PlayerMatchResult): 'win' | 'loss' | 'draw' {
  return result === 'ongoing' || result === 'aborted' ? 'draw' : result
}

export function getTerminationLabel(termination: MatchTermination | undefined): string {
  return {
    '': '未结束',
    checkmate: '将死',
    no_legal_moves: '困毙',
    resign: '认输',
    draw_agreement: '协议和棋',
    threefold_repetition: '三次重复',
  }[termination ?? '']
}

export function getMatchResultStats(matches: MatchSnapshot[]): MatchResultStats {
  const finishedMatches = matches.filter((match) => match.status === 'finished')
  const results = finishedMatches.map(getMatchResult)
  const wins = results.filter((result) => result === 'win').length
  const draws = results.filter((result) => result === 'draw').length
  const losses = results.filter((result) => result === 'loss').length

  return {
    total: matches.length,
    finished: finishedMatches.length,
    wins,
    draws,
    losses,
    winRate: finishedMatches.length > 0
      ? Math.round((wins / finishedMatches.length) * 100)
      : 0,
  }
}

export function getMatchDestination(match: MatchSnapshot): string {
  return isActiveMatch(match.status)
    ? `/match/${match.id}`
    : `/analysis/${match.id}`
}
