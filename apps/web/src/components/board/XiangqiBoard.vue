<script setup lang="ts">
import { onBeforeUnmount, ref, toRef, watch } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'
import { useBoardMotion, type MotionBoardPiece } from '@/composables/useBoardMotion'
import { useMatchStore } from '@/stores/match'
import { useUiStore } from '@/stores/ui'
import { screenPosition } from '@/utils/board'

defineEmits<{ undo: []; resign: [] }>()
const match = useMatchStore()
const ui = useUiStore()
const {
  pieces,
  isAnimating,
  lastSquare,
  arrivalMarker,
} = useBoardMotion({
  fen: toRef(match, 'fen'),
  moves: toRef(match, 'moves'),
  matchId: toRef(match, 'matchId'),
})

const rejectedSquare = ref<{ file: number; rank: number } | null>(null)
let rejectedTimer: ReturnType<typeof setTimeout> | null = null

watch(
  () => match.rejectedMove,
  (rejectedMove) => {
    if (!rejectedMove) return
    if (rejectedTimer) clearTimeout(rejectedTimer)
    rejectedSquare.value = { file: rejectedMove.file, rank: rejectedMove.rank }
    rejectedTimer = setTimeout(() => {
      rejectedSquare.value = null
      rejectedTimer = null
    }, 180)
  },
)

onBeforeUnmount(() => {
  if (rejectedTimer) clearTimeout(rejectedTimer)
})

function handlePieceClick(piece: MotionBoardPiece) {
  if (!match.myTurn || isAnimating.value) return

  if (piece.color === match.playerColor) {
    match.selectPieceAt(piece.file, piece.rank)
    return
  }

  const from = match.selectedPos
  if (from) {
    match.submitMove(from.file, from.rank, piece.file, piece.rank)
  }
}

function move(toFile: number, toRank: number) {
  if (isAnimating.value) return
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

function pieceTrackStyle(piece: MotionBoardPiece) {
  const position = screenPosition(piece.file, piece.rank, match.flipped)
  return {
    '--piece-x': position.left,
    '--piece-y': position.top,
    '--piece-duration': `${piece.motionDurationMs}ms`,
    '--piece-capture-delay': `${Math.max(60, piece.motionDurationMs - 120)}ms`,
  }
}
</script>

<template>
  <div
    class="xiangqi-board"
    :class="{ 'is-animating': isAnimating }"
    :aria-busy="isAnimating"
    :aria-label="`中国象棋棋盘，${match.flipped ? '黑方' : '红方'}视角`"
    @click.self="match.clearSelection"
  >
    <svg class="board-palace" viewBox="0 0 800 900" aria-hidden="true"><g fill="none" stroke="var(--board-line)" stroke-width="2"><path d="M300 0L500 200M500 0L300 200M300 700L500 900M500 700L300 900" /></g></svg>
    <div
      v-for="piece in pieces"
      :key="piece.renderId"
      class="board-piece-track"
      :class="piece.motion"
      :style="pieceTrackStyle(piece)"
    >
      <button
        class="board-piece"
        :class="[
          piece.color,
          {
            selected: match.selectedPos?.file === piece.file && match.selectedPos?.rank === piece.rank,
            last: lastSquare?.file === piece.file && lastSquare?.rank === piece.rank,
            invalid: rejectedSquare?.file === piece.file && rejectedSquare?.rank === piece.rank,
          },
        ]"
        :aria-label="`${piece.color === 'red' ? '红方' : '黑方'}${piece.name}，位置 ${piece.file},${piece.rank}`"
        @click.stop="handlePieceClick(piece)"
      >{{ piece.name }}</button>
    </div>
    <span
      v-if="arrivalMarker"
      :key="arrivalMarker.key"
      class="move-arrival-marker"
      :style="screenPosition(arrivalMarker.file, arrivalMarker.rank, match.flipped)"
      aria-hidden="true"
    />
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
    <button :disabled="!match.allowUndo || match.isFinished || isAnimating" @click="$emit('undo')"><AppIcon name="undo" /><span>悔棋</span></button>
    <button :disabled="isAnimating" @click="flip"><AppIcon name="flip" /><span>翻转</span></button>
    <button @click="toggleSound"><AppIcon name="volume" /><span>音效</span></button>
    <button class="danger-ghost" :disabled="match.isFinished" @click="$emit('resign')"><AppIcon name="flag" /><span>认输</span></button>
  </div>
</template>
