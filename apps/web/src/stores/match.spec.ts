import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useMatchStore } from './match'

describe('match demo store', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it('moves and restores a selected piece', () => {
    const store = useMatchStore()
    const piece = store.pieces.find((item) => item.id === 'r-h1')!
    const original = { file: piece.file, rank: piece.rank }
    store.selectPiece(piece)
    store.moveSelected({ file: 2, rank: 7 })
    expect(piece.file).toBe(2)
    expect(piece.rank).toBe(7)
    expect(store.undo()).toBe(true)
    expect(piece.file).toBe(original.file)
    expect(piece.rank).toBe(original.rank)
  })
})
