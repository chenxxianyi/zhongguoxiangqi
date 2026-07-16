import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useMatchStore } from './match'
import type { MatchSnapshot } from '@/api/contracts'

const push = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push }),
}))

describe('match store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    push.mockClear()
    vi.restoreAllMocks()
  })

  it('selects a local piece and prepares move hints before server validation', () => {
    const store = useMatchStore()
    const piece = store.pieces.find((item) => item.color === 'red' && item.name === '兵')!

    store.selectPieceAt(piece.file, piece.rank)

    expect(store.selectedPos).toEqual({ file: piece.file, rank: piece.rank })
    expect(store.hints.length).toBeGreaterThan(0)

    store.clearSelection()

    expect(store.selectedPos).toBeNull()
    expect(store.hints).toEqual([])
  })

  it('creates a match through the backend contract', async () => {
    const snapshot: MatchSnapshot = {
      id: 'match-1',
      version: 1,
      status: 'active_player_turn',
      playerColor: 'red',
      sideToMove: 'red',
      difficulty: 4,
      aiMode: 'library',
      engine: 'builtin-alpha-beta',
      allowUndo: true,
      initialFen: 'demo',
      fen: 'demo',
      moves: [],
      outcome: 'ongoing',
      drawOffered: false,
      createdAt: new Date(0).toISOString(),
      updatedAt: new Date(0).toISOString(),
    }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify(snapshot), { status: 201, headers: { 'Content-Type': 'application/json' } }),
    )
    const store = useMatchStore()

    await store.createMatch('red', 4, 'library', true)

    const [, init] = fetchMock.mock.calls[0]!
    expect(init?.method).toBe('POST')
    expect(new Headers(init?.headers).get('Idempotency-Key')).toMatch(/^match-create-/)
    expect(JSON.parse(init?.body as string)).toEqual({
      playerColor: 'red',
      difficulty: 4,
      aiMode: 'library',
      allowUndo: true,
    })
    expect(store.matchId).toBe('match-1')
    expect(store.aiMode).toBe('library')
    expect(push).toHaveBeenCalledWith('/match/match-1')
  })
})
