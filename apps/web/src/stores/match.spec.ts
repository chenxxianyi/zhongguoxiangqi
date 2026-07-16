import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useMatchStore } from './match'
import { useUiStore } from './ui'
import { initialFEN } from '@/utils/fen'
import type { MatchSnapshot } from '@/api/contracts'

const push = vi.fn()
const stream = vi.hoisted(() => ({
  handler: null as ((event: any) => void) | null,
  close: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push }),
}))

vi.mock('@/api/stream', () => ({
  connectMatchStream: () => ({
    onEvent(handler: (event: any) => void) {
      stream.handler = handler
    },
    close: stream.close,
  }),
}))

function createSnapshot(overrides: Partial<MatchSnapshot> = {}): MatchSnapshot {
  return {
    id: 'match-1',
    version: 1,
    status: 'active_player_turn',
    playerColor: 'red',
    sideToMove: 'red',
    difficulty: 4,
    aiMode: 'library',
    engine: 'builtin-alpha-beta',
    allowUndo: true,
    initialFen: initialFEN,
    fen: initialFEN,
    moves: [],
    outcome: 'ongoing',
    drawOffered: false,
    createdAt: new Date(0).toISOString(),
    updatedAt: new Date(0).toISOString(),
    ...overrides,
  }
}

describe('match store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    push.mockClear()
    stream.handler = null
    stream.close.mockClear()
    vi.restoreAllMocks()
  })

  it('loads authoritative move hints and caches them for the current version', async () => {
    const snapshot = createSnapshot()
    const fetchMock = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify(snapshot), {
          status: 201,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({
          matchId: snapshot.id,
          matchVersion: snapshot.version,
          sideToMove: 'red',
          moves: [{ move: 'a3a4', from: 'a3', to: 'a4', capture: false }],
        }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
    const store = useMatchStore()
    await store.createMatch('red', 4, 'library', true)
    const piece = store.pieces.find((item) => item.color === 'red' && item.name === '兵')!

    await store.selectPieceAt(piece.file, piece.rank)

    expect(store.selectedPos).toEqual({ file: piece.file, rank: piece.rank })
    expect(store.hints).toEqual([{ file: 0, rank: 5 }])
    expect(fetchMock.mock.calls[1]?.[0]).toBe(
      '/api/v1/matches/match-1/legal-moves?from=a3',
    )

    store.clearSelection()
    await store.selectPieceAt(piece.file, piece.rank)

    expect(store.hints).toEqual([{ file: 0, rank: 5 }])
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })

  it('creates a match through the backend contract', async () => {
    const snapshot = createSnapshot()
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

  it('restores selection and reports an illegal move precisely', async () => {
    const snapshot = createSnapshot()
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify(snapshot), {
          status: 201,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({
          code: 'ILLEGAL_MOVE',
          message: '该着法在当前权威局面中不合法',
        }), {
          status: 422,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
    const store = useMatchStore()
    const ui = useUiStore()
    await store.createMatch('red', 4)

    await store.submitMove(0, 6, 1, 6)

    expect(store.status).toBe('active_player_turn')
    expect(store.selectedPos).toEqual({ file: 0, rank: 6 })
    expect(store.rejectedMove).toMatchObject({ file: 0, rank: 6 })
    expect(ui.toasts.at(-1)?.message).toBe('该着法不合法')
  })

  it('reloads the authoritative snapshot after a version conflict', async () => {
    const snapshot = createSnapshot()
    const updated = createSnapshot({ version: 2, updatedAt: new Date(1).toISOString() })
    const fetchMock = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify(snapshot), {
          status: 201,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({
          code: 'MATCH_VERSION_CONFLICT',
          message: '对局版本已变化，请先获取最新快照',
        }), {
          status: 409,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(updated), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
    const store = useMatchStore()
    const ui = useUiStore()
    await store.createMatch('red', 4)

    await store.submitMove(0, 6, 0, 5)

    expect(fetchMock.mock.calls[2]?.[0]).toBe('/api/v1/matches/match-1')
    expect(store.version).toBe(2)
    expect(store.selectedPos).toBeNull()
    expect(ui.toasts.some((toast) => toast.message === '局面已更新，正在同步')).toBe(true)
  })

  it('leaves an expired match and returns to history', async () => {
    const snapshot = createSnapshot()
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify(snapshot), {
          status: 201,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({
          code: 'MATCH_NOT_FOUND',
          message: '对局不存在或已失效',
        }), {
          status: 404,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
    const store = useMatchStore()
    await store.createMatch('red', 4)

    await store.submitMove(0, 6, 0, 5)

    expect(store.matchId).toBeNull()
    expect(push).toHaveBeenLastCalledWith('/history')
  })

  it('reloads a snapshot instead of applying an event after a version gap', async () => {
    const snapshot = createSnapshot()
    const recovered = createSnapshot({ version: 3, updatedAt: new Date(2).toISOString() })
    const fetchMock = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify(snapshot), {
          status: 201,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(recovered), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
    const store = useMatchStore()
    await store.createMatch('red', 4)

    stream.handler?.({
      eventId: 'event-gap',
      matchId: 'match-1',
      matchVersion: 3,
      type: 'match.ai_move_applied',
      timestamp: new Date(2).toISOString(),
      payload: {},
    })

    await vi.waitFor(() => {
      expect(store.version).toBe(3)
    })
    expect(fetchMock.mock.calls[1]?.[0]).toBe('/api/v1/matches/match-1')
  })
})
