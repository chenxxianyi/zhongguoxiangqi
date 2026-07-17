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
  hints: [] as Array<{ file: number; rank: number }>,
  legalMovesLoading: false,
  selectedPos: { file: 0, rank: 5 } as { file: number; rank: number } | null,
  myTurn: true,
  playerColor: 'red',
  sideToMove: 'red',
  inCheck: false,
  outcome: 'ongoing',
  termination: '',
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
    match.sideToMove = 'red'
    match.inCheck = false
    match.outcome = 'ongoing'
    match.termination = ''
    match.isFinished = false
    match.flipped = false
    match.hints = []
    match.legalMovesLoading = false
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

  it('draws routes and distinguishes empty moves from captures', async () => {
    match.hints = [
      { file: 1, rank: 5 },
      { file: 0, rank: 3 },
    ]
    const wrapper = mount(XiangqiBoard, {
      global: { stubs: { AppIcon: true } },
    })

    expect(wrapper.findAll('.board-route-layer line')).toHaveLength(2)
    expect(wrapper.findAll('.move-hint')).toHaveLength(2)
    expect(wrapper.findAll('.move-hint.capture')).toHaveLength(1)

    await wrapper.find('.move-hint.capture').trigger('click')
    expect(match.submitMove).toHaveBeenCalledWith(0, 5, 0, 3)
  })

  it('keeps the previous move origin and destination visible', () => {
    match.fen = afterFen
    match.moves = [captureMove]
    const wrapper = mount(XiangqiBoard, {
      global: { stubs: { AppIcon: true } },
    })

    expect(wrapper.find('.last-move-marker.from').attributes('style')).toContain('top: 55.05%')
    expect(wrapper.find('.last-move-marker.to').attributes('style')).toContain('top: 34.83%')
  })

  it('warns when the side to move is in check', () => {
    match.fen = '4k4/9/9/9/4R4/9/9/9/9/3K5 b'
    match.selectedPos = null
    match.sideToMove = 'black'
    match.inCheck = true
    const wrapper = mount(XiangqiBoard, {
      global: { stubs: { AppIcon: true } },
    })

    expect(wrapper.find('.board-piece.black.in-check').exists()).toBe(true)
    expect(wrapper.find('.board-check-callout').text()).toContain('将军')
  })

  it('shows a terminal checkmate card with the player result', () => {
    match.fen = '4k4/3RR4/6H2/9/9/9/9/9/9/4K4 b'
    match.selectedPos = null
    match.sideToMove = 'black'
    match.inCheck = true
    match.isFinished = true
    match.outcome = 'red_win'
    match.termination = 'checkmate'
    const wrapper = mount(XiangqiBoard, {
      global: { stubs: { AppIcon: true } },
    })

    expect(wrapper.find('.board-check-callout').exists()).toBe(false)
    expect(wrapper.find('.board-finish-card').text()).toContain('绝杀')
    expect(wrapper.find('.board-finish-card').text()).toContain('你获胜')
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
