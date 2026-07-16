import { mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { MoveRecord } from '@/api/contracts'
import XiangqiBoard from './XiangqiBoard.vue'

const beforeFen = '9/9/9/r8/9/R8/9/9/9/9 w'
const afterFen = '9/9/9/R8/9/9/9/9/9/9 b'

const captureMove: MoveRecord = {
  ply: 1,
  move: 'a4a6',
  side: 'red',
  actor: 'player',
  captured: '车',
  fenBefore: beforeFen,
  fenAfter: afterFen,
  hashAfter: 'capture',
  playedAt: '2026-07-16T00:00:00Z',
}

const match = reactive({
  matchId: 'match-1' as string | null,
  fen: beforeFen,
  moves: [] as MoveRecord[],
  pieces: [
    { color: 'red', name: '车', file: 0, rank: 5 },
    { color: 'black', name: '车', file: 0, rank: 3 },
  ],
  hints: [],
  selectedPos: { file: 0, rank: 5 },
  myTurn: true,
  playerColor: 'red',
  flipped: false,
  allowUndo: true,
  isFinished: false,
  soundEnabled: true,
  rejectedMove: null as { id: number; file: number; rank: number } | null,
  selectPieceAt: vi.fn(),
  submitMove: vi.fn(),
  clearSelection: vi.fn(),
})

vi.mock('@/stores/match', () => ({
  useMatchStore: () => match,
}))

vi.mock('@/stores/ui', () => ({
  useUiStore: () => ({ showToast: vi.fn() }),
}))

describe('XiangqiBoard', () => {
  beforeEach(() => {
    match.matchId = 'match-1'
    match.fen = beforeFen
    match.moves = []
    match.selectedPos = { file: 0, rank: 5 }
    match.myTurn = true
    match.playerColor = 'red'
    match.flipped = false
    match.rejectedMove = null
    match.selectPieceAt.mockClear()
    match.submitMove.mockClear()
  })

  it('submits a capture when an opponent piece is clicked after selecting a piece', async () => {
    const wrapper = mount(XiangqiBoard, {
      global: { stubs: { AppIcon: true } },
    })

    await wrapper.find('.board-piece.black').trigger('click')

    expect(match.submitMove).toHaveBeenCalledWith(0, 5, 0, 3)
    expect(match.selectPieceAt).not.toHaveBeenCalled()
  })

  it('switches selection when a friendly piece is clicked', async () => {
    const wrapper = mount(XiangqiBoard, {
      global: { stubs: { AppIcon: true } },
    })

    await wrapper.find('.board-piece.red').trigger('click')

    expect(match.selectPieceAt).toHaveBeenCalledWith(0, 5)
    expect(match.submitMove).not.toHaveBeenCalled()
  })

  it('keeps the moving piece mounted while a captured piece exits', async () => {
    vi.useFakeTimers()
    const wrapper = mount(XiangqiBoard, {
      global: { stubs: { AppIcon: true } },
    })
    const movingPiece = wrapper.find('.board-piece.red').element

    match.moves = [captureMove]
    match.fen = afterFen
    await nextTick()

    expect(wrapper.find('.board-piece-track.moving').exists()).toBe(true)
    expect(wrapper.find('.board-piece-track.captured').exists()).toBe(true)
    expect(wrapper.find('.board-piece.red').element).toBe(movingPiece)
    expect(wrapper.findAll('.board-piece')).toHaveLength(2)

    await vi.advanceTimersByTimeAsync(400)
    await nextTick()

    expect(wrapper.findAll('.board-piece')).toHaveLength(1)
    expect(wrapper.find('.board-piece-track.moving').exists()).toBe(false)
    vi.useRealTimers()
  })

  it('animates a capture in reverse and restores the captured piece on undo', async () => {
    vi.useFakeTimers()
    match.fen = afterFen
    match.moves = [captureMove]
    const wrapper = mount(XiangqiBoard, {
      global: { stubs: { AppIcon: true } },
    })

    match.moves = []
    match.fen = beforeFen
    await nextTick()

    expect(wrapper.find('.board-piece-track.moving').exists()).toBe(true)
    expect(wrapper.find('.board-piece-track.restored').exists()).toBe(true)

    await vi.advanceTimersByTimeAsync(300)
    await nextTick()

    expect(wrapper.findAll('.board-piece')).toHaveLength(2)
    expect(wrapper.find('.board-piece.red').attributes('aria-label')).toContain('位置 0,5')
    expect(wrapper.find('.board-piece.black').attributes('aria-label')).toContain('位置 0,3')
    vi.useRealTimers()
  })
})
