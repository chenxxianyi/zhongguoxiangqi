import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import XiangqiBoard from './XiangqiBoard.vue'

const match = {
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
  selectPieceAt: vi.fn(),
  submitMove: vi.fn(),
  clearSelection: vi.fn(),
}

vi.mock('@/stores/match', () => ({
  useMatchStore: () => match,
}))

vi.mock('@/stores/ui', () => ({
  useUiStore: () => ({ showToast: vi.fn() }),
}))

describe('XiangqiBoard', () => {
  beforeEach(() => {
    match.selectedPos = { file: 0, rank: 5 }
    match.myTurn = true
    match.playerColor = 'red'
    match.selectPieceAt.mockClear()
    match.submitMove.mockClear()
  })

  it('submits a capture when an opponent piece is clicked after selecting a piece', async () => {
    const wrapper = mount(XiangqiBoard, {
      global: { stubs: { AppIcon: true } },
    })

    await wrapper.findAll('.board-piece')[1]!.trigger('click')

    expect(match.submitMove).toHaveBeenCalledWith(0, 5, 0, 3)
    expect(match.selectPieceAt).not.toHaveBeenCalled()
  })

  it('switches selection when a friendly piece is clicked', async () => {
    const wrapper = mount(XiangqiBoard, {
      global: { stubs: { AppIcon: true } },
    })

    await wrapper.findAll('.board-piece')[0]!.trigger('click')

    expect(match.selectPieceAt).toHaveBeenCalledWith(0, 5)
    expect(match.submitMove).not.toHaveBeenCalled()
  })
})
