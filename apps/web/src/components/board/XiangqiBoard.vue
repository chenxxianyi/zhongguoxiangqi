<script setup lang="ts">
import AppIcon from '@/components/common/AppIcon.vue'
import { useMatchStore } from '@/stores/match'
import { useUiStore } from '@/stores/ui'
import { screenPosition } from '@/utils/board'

defineEmits<{ undo: []; resign: [] }>()
const match = useMatchStore()
const ui = useUiStore()
function select(pieceId: string) { const piece = match.pieces.find((item) => item.id === pieceId); if (piece) match.selectPiece(piece) }
function move(file: number, rank: number) { const piece = match.moveSelected({ file, rank }); if (piece) ui.showToast(`已演示落子：${piece.name}移动到新位置`) }
function flip() { match.flipped = !match.flipped; ui.showToast(match.flipped ? '已切换为黑方视角' : '已切换为红方视角') }
function toggleSound() { match.soundEnabled = !match.soundEnabled; ui.showToast(match.soundEnabled ? '落子音效已开启' : '落子音效已关闭') }
</script>

<template>
  <div class="xiangqi-board" :aria-label="`中国象棋棋盘，${match.flipped ? '黑方' : '红方'}视角`" @click.self="match.clearSelection">
    <svg class="board-palace" viewBox="0 0 800 900" aria-hidden="true"><g fill="none" stroke="var(--board-line)" stroke-width="2"><path d="M300 0L500 200M500 0L300 200M300 700L500 900M500 700L300 900" /></g></svg>
    <button v-for="piece in match.pieces" :key="piece.id" class="board-piece" :class="[piece.color, { selected: match.selectedId === piece.id, last: piece.last }]" :style="screenPosition(piece.file, piece.rank, match.flipped)" :aria-label="`${piece.color === 'red' ? '红方' : '黑方'}${piece.name}，位置 ${piece.file},${piece.rank}`" @click.stop="select(piece.id)">{{ piece.name }}</button>
    <button v-for="hint in match.hints" :key="`${hint.file}-${hint.rank}`" class="move-hint" :style="screenPosition(hint.file, hint.rank, match.flipped)" :aria-label="`演示移动到 ${hint.file},${hint.rank}`" @click.stop="move(hint.file, hint.rank)" />
  </div>
  <slot />
  <div class="board-toolbar" aria-label="棋盘工具"><button @click="$emit('undo')"><AppIcon name="undo" /><span>悔棋</span></button><button @click="flip"><AppIcon name="flip" /><span>翻转</span></button><button @click="toggleSound"><AppIcon name="volume" /><span>音效</span></button><button class="danger-ghost" @click="$emit('resign')"><AppIcon name="flag" /><span>认输</span></button></div>
</template>
