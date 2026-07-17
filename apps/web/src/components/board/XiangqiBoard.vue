<script setup lang="ts">
import { computed, onBeforeUnmount, ref, toRef, watch } from 'vue'
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
  lastMoveSquares,
  arrivalMarker,
} = useBoardMotion({
  fen: toRef(match, 'fen'),
  moves: toRef(match, 'moves'),
  matchId: toRef(match, 'matchId'),
})

const rejectedSquare = ref<{ file: number; rank: number } | null>(null)
let rejectedTimer: ReturnType<typeof setTimeout> | null = null

const selectedScreenPoint = computed(() => {
  if (!match.selectedPos) return null
  return screenPoint(match.selectedPos.file, match.selectedPos.rank)
})

const routeLines = computed(() => {
  if (!selectedScreenPoint.value || match.legalMovesLoading) return []
  return match.hints.map((hint) => ({
    ...hint,
    from: selectedScreenPoint.value!,
    to: screenPoint(hint.file, hint.rank),
    capture: isOccupied(hint.file, hint.rank),
  }))
})

const boardAriaLabel = computed(() => {
  const orientation = match.flipped ? '黑方' : '红方'
  if (match.termination === 'checkmate') return `中国象棋棋盘，${orientation}视角，对局已将死`
  if (match.inCheck) return `中国象棋棋盘，${orientation}视角，${match.sideToMove === match.playerColor ? '你被将军' : '对方被将军'}`
  if (match.legalMovesLoading) return `中国象棋棋盘，${orientation}视角，正在获取合法落点`
  if (match.selectedPos) return `中国象棋棋盘，${orientation}视角，已显示 ${match.hints.length} 个合法落点`
  return `中国象棋棋盘，${orientation}视角`
})

const checkmatePresentation = computed(() => {
  if (!match.isFinished || match.termination !== 'checkmate') return null
  const winner = match.outcome === 'red_win' ? '红方' : '黑方'
  const loser = winner === '红方' ? '黑方' : '红方'
  const playerWon = (
    match.outcome === 'red_win' && match.playerColor === 'red'
  ) || (
    match.outcome === 'black_win' && match.playerColor === 'black'
  )
  return {
    title: playerWon ? '绝杀' : '被将死',
    result: playerWon ? '你获胜' : '你惜败',
    detail: `${winner}胜 · ${loser}无合法应将`,
  }
})

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

function screenPoint(file: number, rank: number) {
  const position = screenPosition(file, rank, match.flipped)
  return {
    x: Number.parseFloat(position.left),
    y: Number.parseFloat(position.top),
  }
}

function isOccupied(file: number, rank: number) {
  return pieces.value.some((piece) =>
    piece.motion !== 'captured' && piece.file === file && piece.rank === rank,
  )
}

function isCheckedGeneral(piece: MotionBoardPiece) {
  return match.inCheck
    && piece.color === match.sideToMove
    && (piece.name === '帅' || piece.name === '将')
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
    :class="{ 'is-animating': isAnimating, 'has-selection': match.selectedPos }"
    :aria-busy="isAnimating"
    :aria-label="boardAriaLabel"
    @click="match.clearSelection"
  >
    <svg class="board-palace" viewBox="0 0 800 900" aria-hidden="true"><g fill="none" stroke="var(--board-line)" stroke-width="2"><path d="M300 0L500 200M500 0L300 200M300 700L500 900M500 700L300 900" /></g></svg>
    <svg
      v-if="routeLines.length"
      class="board-route-layer"
      viewBox="0 0 100 100"
      preserveAspectRatio="none"
      aria-hidden="true"
    >
      <line
        v-for="route in routeLines"
        :key="`route-${route.file}-${route.rank}`"
        :class="{ capture: route.capture }"
        :x1="route.from.x"
        :y1="route.from.y"
        :x2="route.to.x"
        :y2="route.to.y"
        vector-effect="non-scaling-stroke"
      />
    </svg>
    <span
      v-if="lastMoveSquares"
      class="last-move-marker from"
      :style="screenPosition(lastMoveSquares.from.file, lastMoveSquares.from.rank, match.flipped)"
      aria-hidden="true"
    />
    <span
      v-if="lastMoveSquares"
      class="last-move-marker to"
      :style="screenPosition(lastMoveSquares.to.file, lastMoveSquares.to.rank, match.flipped)"
      aria-hidden="true"
    />
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
            last: lastMoveSquares?.to.file === piece.file && lastMoveSquares?.to.rank === piece.rank,
            loading: match.legalMovesLoading && match.selectedPos?.file === piece.file && match.selectedPos?.rank === piece.rank,
            'in-check': isCheckedGeneral(piece),
            invalid: rejectedSquare?.file === piece.file && rejectedSquare?.rank === piece.rank,
          },
        ]"
        :aria-label="`${piece.color === 'red' ? '红方' : '黑方'}${piece.name}，位置 ${piece.file},${piece.rank}`"
        :aria-pressed="match.selectedPos?.file === piece.file && match.selectedPos?.rank === piece.rank"
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
      :class="{ capture: isOccupied(hint.file, hint.rank) }"
      :style="screenPosition(hint.file, hint.rank, match.flipped)"
      :aria-label="`${isOccupied(hint.file, hint.rank) ? '吃子并移动' : '移动'}到 ${hint.file},${hint.rank}`"
      @click.stop="move(hint.file, hint.rank)"
    />
    <span class="sr-only" aria-live="polite">
      {{ match.legalMovesLoading ? '正在获取合法落点' : match.selectedPos ? `找到 ${match.hints.length} 个合法落点` : '' }}
    </span>
    <div
      v-if="match.inCheck && !match.isFinished && !isAnimating"
      class="board-check-callout"
      role="status"
      aria-live="assertive"
    >
      <span class="board-state-seal">将</span>
      <span>
        <strong>{{ match.sideToMove === match.playerColor ? '你被将军' : '将军' }}</strong>
        <small>{{ match.sideToMove === match.playerColor ? '请选择合法着法应将' : '对方正在应将' }}</small>
      </span>
    </div>
    <section
      v-if="checkmatePresentation && !isAnimating"
      class="board-finish-overlay"
      role="alert"
      aria-live="assertive"
    >
      <div class="board-finish-card">
        <span class="board-state-seal mate">杀</span>
        <span class="board-finish-kicker">CHECKMATE · 对局结束</span>
        <h2>{{ checkmatePresentation.title }}</h2>
        <strong>{{ checkmatePresentation.result }}</strong>
        <p>{{ checkmatePresentation.detail }}</p>
      </div>
    </section>
  </div>
  <slot />
  <div class="board-toolbar" aria-label="棋盘工具">
    <button :disabled="!match.allowUndo || match.isFinished || isAnimating" @click="$emit('undo')"><AppIcon name="undo" /><span>悔棋</span></button>
    <button :disabled="isAnimating" @click="flip"><AppIcon name="flip" /><span>翻转</span></button>
    <button @click="toggleSound"><AppIcon name="volume" /><span>音效</span></button>
    <button class="danger-ghost" :disabled="match.isFinished" @click="$emit('resign')"><AppIcon name="flag" /><span>认输</span></button>
  </div>
</template>
