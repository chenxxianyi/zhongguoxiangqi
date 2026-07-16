<script setup lang="ts">
import AppIcon from '@/components/common/AppIcon.vue'
import { useMatchStore } from '@/stores/match'
import { useUiStore } from '@/stores/ui'
import { screenPosition } from '@/utils/board'

defineEmits<{ undo: []; resign: [] }>()
const match = useMatchStore()
const ui = useUiStore()

function handlePieceClick(file: number, rank: number) {
  const piece = match.pieces.find((p) => p.file === file && p.rank === rank)
  if (!piece) return

  if (!match.myTurn) return

  if (piece.color === match.playerColor) {
    match.selectPieceAt(file, rank)
    return
  }

  const from = match.selectedPos
  if (from) {
    match.submitMove(from.file, from.rank, file, rank)
  }
}

function move(toFile: number, toRank: number) {
  const from = match.selectedPos
  if (!from) return
  match.submitMove(from.file, from.rank, toFile, toRank)
}

function flip() {
  match.flipped = !match.flipped
  ui.showToast(match.flipped ? '已切换为黑方视角' : '已切换为红方视角')
}

function toggleSound() {
  match.soundEnabled = !match.soundEnabled
  ui.showToast(match.soundEnabled ? '落子音效已开启' : '落子音效已关闭')
}
</script>

<template>
  <div class="xiangqi-board" :aria-label="`中国象棋棋盘，${match.flipped ? '黑方' : '红方'}视角`" @click.self="match.clearSelection">
    <svg class="board-palace" viewBox="0 0 800 900" aria-hidden="true"><g fill="none" stroke="var(--board-line)" stroke-width="2"><path d="M300 0L500 200M500 0L300 200M300 700L500 900M500 700L300 900" /></g></svg>
    <button
      v-for="piece in match.pieces"
      :key="`${piece.color}-${piece.name}-${piece.file}-${piece.rank}`"
      class="board-piece"
      :class="[
        piece.color,
        {
          selected: match.selectedPos?.file === piece.file && match.selectedPos?.rank === piece.rank,
          last: piece.last,
        },
      ]"
      :style="screenPosition(piece.file, piece.rank, match.flipped)"
      :aria-label="`${piece.color === 'red' ? '红方' : '黑方'}${piece.name}，位置 ${piece.file},${piece.rank}`"
      @click.stop="handlePieceClick(piece.file, piece.rank)"
    >{{ piece.name }}</button>
    <button
      v-for="hint in match.hints"
      :key="`hint-${hint.file}-${hint.rank}`"
      class="move-hint"
      :style="screenPosition(hint.file, hint.rank, match.flipped)"
      :aria-label="`移动到 ${hint.file},${hint.rank}`"
      @click.stop="move(hint.file, hint.rank)"
    />
  </div>
  <slot />
  <div class="board-toolbar" aria-label="棋盘工具">
    <button :disabled="!match.allowUndo || match.isFinished" @click="$emit('undo')"><AppIcon name="undo" /><span>悔棋</span></button>
    <button @click="flip"><AppIcon name="flip" /><span>翻转</span></button>
    <button @click="toggleSound"><AppIcon name="volume" /><span>音效</span></button>
    <button class="danger-ghost" :disabled="match.isFinished" @click="$emit('resign')"><AppIcon name="flag" /><span>认输</span></button>
  </div>
</template>
