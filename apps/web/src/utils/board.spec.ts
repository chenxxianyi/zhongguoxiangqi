import { describe, expect, it } from 'vitest'
import { fromICCSSquare, screenPosition, toICCSSquare } from '@/utils/board'

describe('board display coordinates', () => {
  it('maps red and black orientations consistently', () => {
    expect(screenPosition(0, 0, false)).toEqual({ left: '4.5%', top: '4.5%' })
    expect(screenPosition(0, 0, true)).toEqual({ left: '95.5%', top: '95.49%' })
  })
})

describe('ICCS board coordinates', () => {
  it('round trips UI coordinates without using local move rules', () => {
    expect(toICCSSquare(0, 6)).toBe('a3')
    expect(fromICCSSquare('a3')).toEqual({ file: 0, rank: 6 })
    expect(fromICCSSquare('invalid')).toBeNull()
  })
})
