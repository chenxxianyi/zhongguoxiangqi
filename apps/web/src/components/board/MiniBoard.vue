<script setup lang="ts">
import type { BoardPiece } from '@/types/xiangqi'
import { screenPosition } from '@/utils/board'

withDefaults(defineProps<{ pieces?: BoardPiece[] }>(), { pieces: () => [] })
</script>

<template>
  <div class="mini-board">
    <span
      v-for="piece in pieces"
      :key="piece.id"
      class="mini-piece"
      :class="piece.color"
      :style="screenPosition(piece.file, piece.rank, false)"
    >
      {{ piece.name }}
    </span>
  </div>
</template>
