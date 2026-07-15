import { describe, expect, it } from 'vitest'
import { candidateMoves, screenPosition } from '@/utils/board'
import type { BoardPiece } from '@/types/xiangqi'

describe('board display coordinates', () => {
  it('maps red and black orientations consistently', () => {
    expect(screenPosition(0, 0, false)).toEqual({ left: '4.5%', top: '4.5%' })
    expect(screenPosition(0, 0, true)).toEqual({ left: '95.5%', top: '95.49%' })
  })
})

describe('demo candidate moves', () => {
  it('keeps every hint inside the 9 x 10 board', () => {
    const piece: BoardPiece = { id: 'horse', color: 'red', name: '马', file: 1, rank: 9 }
    for (const move of candidateMoves(piece)) {
      expect(move.file).toBeGreaterThanOrEqual(0)
      expect(move.file).toBeLessThanOrEqual(8)
      expect(move.rank).toBeGreaterThanOrEqual(0)
      expect(move.rank).toBeLessThanOrEqual(9)
    }
  })
})
