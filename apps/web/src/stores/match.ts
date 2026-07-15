import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { initialPieces } from '@/data/demo'
import { candidateMoves } from '@/utils/board'
import type { BoardPiece, BoardSquare } from '@/types/xiangqi'

export const useMatchStore = defineStore('match', () => {
  const pieces = ref<BoardPiece[]>(structuredClone(initialPieces))
  const flipped = ref(false)
  const soundEnabled = ref(true)
  const selectedId = ref<string | null>(null)
  const hints = ref<BoardSquare[]>([])
  const history = ref<Array<{ id: string; file: number; rank: number; lastIds: string[] }>>([])

  const selectedPiece = computed(() => pieces.value.find((piece) => piece.id === selectedId.value) ?? null)

  function selectPiece(piece: BoardPiece) {
    selectedId.value = piece.id
    hints.value = candidateMoves(piece)
  }

  function clearSelection() {
    selectedId.value = null
    hints.value = []
  }

  function moveSelected(square: BoardSquare) {
    const piece = selectedPiece.value
    if (!piece) return null
    history.value.push({ id: piece.id, file: piece.file, rank: piece.rank, lastIds: pieces.value.filter((item) => item.last).map((item) => item.id) })
    pieces.value.forEach((item) => { item.last = false })
    piece.file = square.file
    piece.rank = square.rank
    piece.last = true
    clearSelection()
    return piece
  }

  function undo() {
    const previous = history.value.pop()
    if (!previous) return false
    const piece = pieces.value.find((item) => item.id === previous.id)
    if (!piece) return false
    pieces.value.forEach((item) => { item.last = previous.lastIds.includes(item.id) })
    piece.file = previous.file
    piece.rank = previous.rank
    clearSelection()
    return true
  }

  function reset() {
    pieces.value = structuredClone(initialPieces)
    history.value = []
    clearSelection()
  }

  return { pieces, flipped, soundEnabled, selectedId, hints, selectedPiece, selectPiece, clearSelection, moveSelected, undo, reset }
})
