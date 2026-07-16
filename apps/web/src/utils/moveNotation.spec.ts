import { describe, expect, it } from 'vitest'
import { formatMoveList, iccsToDisplay } from './moveNotation'

describe('move notation', () => {
  it('uses the position before each move to resolve the moving piece', () => {
    const redFen = '4k4/9/9/9/9/9/9/9/9/R3K4 w'
    const blackFen = 'r3k4/9/9/9/9/9/9/9/9/4K4 b'
    const rows = formatMoveList([
      {
        ply: 1,
        move: 'a0a1',
        side: 'red',
        actor: 'player',
        fenBefore: redFen,
      },
      {
        ply: 2,
        move: 'a9a8',
        side: 'black',
        actor: 'ai',
        fenBefore: blackFen,
      },
    ], '9/9/9/9/4k4/4K4/9/9/9/9 w')

    expect(rows[0]?.[1]).toMatch(/^车/)
    expect(rows[0]?.[2]).toMatch(/^车/)
  })

  it('uses the same starting position for actual and recommended moves', () => {
    const fenBefore = '4k4/9/9/9/9/9/9/9/9/R3K4 w'

    expect(iccsToDisplay('a0a1', fenBefore, 'red')).toMatch(/^车/)
    expect(iccsToDisplay('a0b0', fenBefore, 'red')).toMatch(/^车/)
  })
})
