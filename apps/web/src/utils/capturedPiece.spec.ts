import { describe, expect, it } from 'vitest'
import { getCapturedPieceLabel } from './capturedPiece'

describe('captured piece labels', () => {
  it('maps backend piece identifiers to Chinese chess labels', () => {
    expect(getCapturedPieceLabel('black_rook')).toBe('车')
    expect(getCapturedPieceLabel('red_advisor')).toBe('仕')
    expect(getCapturedPieceLabel('black_pawn')).toBe('卒')
  })

  it('keeps already localized and empty values stable', () => {
    expect(getCapturedPieceLabel('炮')).toBe('炮')
    expect(getCapturedPieceLabel(undefined)).toBe('')
  })
})
