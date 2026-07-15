<script setup lang="ts">
import { computed } from 'vue'
import type { BoardPiece } from '@/types/xiangqi'
import { screenPosition } from '@/utils/board'

const props = withDefaults(defineProps<{ pieces?: BoardPiece[] }>(), { pieces: () => [] })
const displayPieces = computed(() => props.pieces.length ? props.pieces : [
  { id: 'mini-bk', color: 'black', name: '将', file: 4, rank: 0 },
  { id: 'mini-br', color: 'black', name: '车', file: 0, rank: 0 },
  { id: 'mini-bh', color: 'black', name: '马', file: 7, rank: 2 },
  { id: 'mini-bc', color: 'black', name: '炮', file: 3, rank: 3 },
  { id: 'mini-bp', color: 'black', name: '卒', file: 4, rank: 4 },
  { id: 'mini-rk', color: 'red', name: '帅', file: 4, rank: 9 },
  { id: 'mini-rr', color: 'red', name: '车', file: 8, rank: 8 },
  { id: 'mini-rh', color: 'red', name: '马', file: 2, rank: 7 },
  { id: 'mini-rc', color: 'red', name: '炮', file: 5, rank: 6 },
  { id: 'mini-rp', color: 'red', name: '兵', file: 4, rank: 5 },
] as BoardPiece[])
</script>

<template>
  <div class="mini-board">
    <span v-for="piece in displayPieces" :key="piece.id" class="mini-piece" :class="piece.color" :style="screenPosition(piece.file, piece.rank, false)">{{ piece.name }}</span>
  </div>
</template>
