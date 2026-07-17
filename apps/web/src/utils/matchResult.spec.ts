import { describe, expect, it } from 'vitest'
import type { MatchSnapshot } from '@/api/contracts'
import {
  getMatchDestination,
  getMatchResult,
  getMatchResultStats,
  getTerminationLabel,
  getPlayerResultClass,
  getPlayerResult,
  isActiveMatch,
} from './matchResult'

function createMatch(overrides: Partial<MatchSnapshot> = {}): MatchSnapshot {
  return {
    id: 'match-1',
    version: 1,
    status: 'finished',
    playerColor: 'red',
    sideToMove: 'red',
    difficulty: 6,
    aiMode: 'standard',
    engine: 'builtin',
    allowUndo: true,
    initialFen: 'initial',
    fen: 'final',
    moves: [],
    outcome: 'red_win',
    drawOffered: false,
    createdAt: '2026-07-16T00:00:00Z',
    updatedAt: '2026-07-16T00:00:00Z',
    ...overrides,
  }
}

describe('match result helpers', () => {
  it('treats a black-side victory as a player win', () => {
    expect(getPlayerResult('black_win', 'black')).toBe('win')
    expect(getMatchResult(createMatch({
      playerColor: 'black',
      outcome: 'black_win',
    }))).toBe('win')
  })

  it('keeps unfinished matches out of loss statistics', () => {
    const matches = [
      createMatch(),
      createMatch({ id: 'draw', outcome: 'draw' }),
      createMatch({
        id: 'active',
        status: 'active_player_turn',
        outcome: 'ongoing',
      }),
    ]

    expect(getMatchResultStats(matches)).toEqual({
      total: 3,
      finished: 2,
      wins: 1,
      draws: 1,
      losses: 0,
      winRate: 50,
    })
  })

  it('routes active matches back to the board and completed matches to analysis', () => {
    const active = createMatch({ status: 'recoverable_error', outcome: 'ongoing' })
    const finished = createMatch()

    expect(isActiveMatch(active.status)).toBe(true)
    expect(getMatchDestination(active)).toBe('/match/match-1')
    expect(getMatchDestination(finished)).toBe('/analysis/match-1')
  })

  it('renders aborted matches as a neutral terminal result', () => {
    const aborted = createMatch({ status: 'aborted', outcome: 'ongoing' })

    expect(getMatchResult(aborted)).toBe('aborted')
    expect(getPlayerResultClass(getMatchResult(aborted))).toBe('draw')
    expect(getMatchDestination(aborted)).toBe('/analysis/match-1')
  })

  it('uses precise Chinese labels for terminal reasons', () => {
    expect(getTerminationLabel('checkmate')).toBe('将死')
    expect(getTerminationLabel('no_legal_moves')).toBe('困毙')
    expect(getTerminationLabel(undefined)).toBe('未结束')
  })
})
