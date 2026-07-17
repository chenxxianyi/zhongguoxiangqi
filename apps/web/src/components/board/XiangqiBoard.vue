<script setup lang="ts">
import { computed, onBeforeUnmount, ref, toRef, watch } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'
import { useBoardMotion, type MotionBoardPiece } from '@/composables/useBoardMotion'
import { useBoardSound } from '@/composables/useBoardSound'
import { useMatchStore } from '@/stores/match'
import { useUiStore } from '@/stores/ui'
import { screenPosition } from '@/utils/board'
import { getCapturedPieceLabel } from '@/utils/capturedPiece'

const emit = defineEmits<{ undo: []; resign: []; restart: []; review: [] }>()
const match = useMatchStore()
const ui = useUiStore()
const boardSound = useBoardSound()
const {
  pieces,
  isAnimating,
  lastMoveSquares,
  arrivalMarker,
  captureMarker,
  motionPhase,
  reducedMotion,
} = useBoardMotion({
  fen: toRef(match, 'fen'),
  moves: toRef(match, 'moves'),
  matchId: toRef(match, 'matchId'),
})

const rejectedSquare = ref<{ file: number; rank: number } | null>(null)
const boardElement = ref<HTMLElement | null>(null)
const focusedBoardTarget = ref<string | null>(null)
const activeRouteTarget = ref<{ file: number; rank: number } | null>(null)
let rejectedTimer: ReturnType<typeof setTimeout> | null = null
let audibleMatchId: string | null = null
let lastAudiblePly = 0

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
    active: activeRouteTarget.value?.file === hint.file && activeRouteTarget.value?.rank === hint.rank,
  }))
})

const lastMoveRoute = computed(() => {
  if (!lastMoveSquares.value) return null
  return {
    from: screenPoint(lastMoveSquares.value.from.file, lastMoveSquares.value.from.rank),
    to: screenPoint(lastMoveSquares.value.to.file, lastMoveSquares.value.to.rank),
  }
})

const activeFocusTarget = computed(() => {
  const available = new Set([
    ...pieces.value
      .filter((piece) => piece.motion !== 'captured')
      .map((piece) => `piece-${piece.renderId}`),
    ...match.hints.map((hint) => `hint-${hint.file}-${hint.rank}`),
  ])
  if (focusedBoardTarget.value && available.has(focusedBoardTarget.value)) return focusedBoardTarget.value
  const selected = pieces.value.find((piece) =>
    piece.motion !== 'captured'
    && piece.file === match.selectedPos?.file
    && piece.rank === match.selectedPos?.rank,
  )
  if (selected) return `piece-${selected.renderId}`
  const ownPiece = pieces.value.find((piece) => piece.motion !== 'captured' && piece.color === match.playerColor)
  const fallback = ownPiece ?? pieces.value.find((piece) => piece.motion !== 'captured')
  return fallback ? `piece-${fallback.renderId}` : ''
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

const capturePresentation = computed(() => {
  if (!captureMarker.value) return null
  const label = getCapturedPieceLabel(captureMarker.value.captured)
  if (!label) return null
  const point = screenPoint(captureMarker.value.file, captureMarker.value.rank)
  return {
    label,
    announcement: captureMarker.value.actor === 'player'
      ? `你吃掉了对方的${label}`
      : `对方吃掉了你的${label}`,
    placement: {
      below: point.y < 18,
      'edge-left': point.x < 12,
      'edge-right': point.x > 88,
    },
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

watch(
  () => match.selectedPos,
  () => {
    activeRouteTarget.value = null
  },
)

watch(
  () => [match.matchId, match.moves.at(-1)?.ply ?? 0] as const,
  ([matchId, latestPly]) => {
    if (!matchId || matchId !== audibleMatchId) {
      audibleMatchId = matchId
      lastAudiblePly = latestPly
      return
    }
    if (latestPly <= lastAudiblePly) {
      lastAudiblePly = latestPly
      return
    }
    const latestMove = match.moves.at(-1)
    lastAudiblePly = latestPly
    if (latestMove && match.soundEnabled) {
      boardSound.playMove(latestMove)
      if (!reducedMotion.value && (latestMove.captured || latestMove.givesCheck)) {
        navigator.vibrate?.(latestMove.givesCheck ? [12, 22, 12] : 10)
      }
    }
  },
)

onBeforeUnmount(() => {
  if (rejectedTimer) clearTimeout(rejectedTimer)
  boardSound.dispose()
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

function handleBoardArrow(event: KeyboardEvent) {
  if (!['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(event.key)) return
  const current = event.currentTarget
  if (!(current instanceof HTMLElement) || !boardElement.value) return
  const file = Number(current.dataset.boardFile)
  const rank = Number(current.dataset.boardRank)
  if (!Number.isFinite(file) || !Number.isFinite(rank)) return

  event.preventDefault()
  const from = screenPoint(file, rank)
  const candidates = [...boardElement.value.querySelectorAll<HTMLElement>('[data-board-focus]')]
    .filter((element) => element !== current)
    .map((element) => {
      const point = screenPoint(Number(element.dataset.boardFile), Number(element.dataset.boardRank))
      return { element, dx: point.x - from.x, dy: point.y - from.y }
    })
    .filter(({ dx, dy }) => {
      if (event.key === 'ArrowLeft') return dx < -0.1
      if (event.key === 'ArrowRight') return dx > 0.1
      if (event.key === 'ArrowUp') return dy < -0.1
      return dy > 0.1
    })
    .sort((left, right) => {
      const leftScore = event.key === 'ArrowLeft' || event.key === 'ArrowRight'
        ? Math.abs(left.dx) + Math.abs(left.dy) * 3
        : Math.abs(left.dy) + Math.abs(left.dx) * 3
      const rightScore = event.key === 'ArrowLeft' || event.key === 'ArrowRight'
        ? Math.abs(right.dx) + Math.abs(right.dy) * 3
        : Math.abs(right.dy) + Math.abs(right.dx) * 3
      return leftScore - rightScore
    })
  candidates[0]?.element.focus()
}

function move(toFile: number, toRank: number) {
  if (isAnimating.value) return
  const from = match.selectedPos
  if (!from) return
  match.submitMove(from.file, from.rank, toFile, toRank)
}

function activateRoute(file: number, rank: number) {
  activeRouteTarget.value = { file, rank }
}

function clearActiveRoute(file: number, rank: number) {
  if (activeRouteTarget.value?.file === file && activeRouteTarget.value?.rank === rank) {
    activeRouteTarget.value = null
  }
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
  if (match.soundEnabled) boardSound.playEnabled()
  ui.showToast(match.soundEnabled ? '落子音效已开启' : '落子音效已关闭')
}

function pieceTrackStyle(piece: MotionBoardPiece) {
  const position = screenPosition(piece.file, piece.rank, match.flipped)
  return {
    '--piece-x': position.left,
    '--piece-y': position.top,
    '--piece-duration': `${piece.motionDurationMs}ms`,
    '--piece-capture-delay': `${Math.max(50, piece.motionDurationMs - 150)}ms`,
  }
}
</script>

<template>
  <div
    ref="boardElement"
    class="xiangqi-board"
    :class="{
      'is-animating': isAnimating,
      'has-selection': match.selectedPos,
      'has-route-focus': activeRouteTarget,
      'is-in-check': match.inCheck && !match.isFinished,
      'is-finished': match.isFinished,
    }"
    :data-motion-phase="motionPhase"
    :aria-busy="isAnimating"
    :aria-label="boardAriaLabel"
    aria-describedby="board-keyboard-help"
    @click="match.clearSelection"
  >
    <div class="board-river" aria-hidden="true"><span>楚河</span><i>棋境</i><span>汉界</span></div>
    <svg class="board-palace" viewBox="0 0 800 900" aria-hidden="true"><g fill="none" stroke="var(--board-line)" stroke-width="2"><path d="M300 0L500 200M500 0L300 200M300 700L500 900M500 700L300 900" /></g></svg>
    <svg
      v-if="lastMoveRoute"
      :key="`last-route-${match.moves.length}`"
      class="last-move-route-layer"
      viewBox="0 0 100 100"
      preserveAspectRatio="none"
      aria-hidden="true"
    >
      <line
        :x1="lastMoveRoute.from.x"
        :y1="lastMoveRoute.from.y"
        :x2="lastMoveRoute.to.x"
        :y2="lastMoveRoute.to.y"
        pathLength="1"
        vector-effect="non-scaling-stroke"
      />
    </svg>
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
        :class="{
          capture: route.capture,
          active: route.active,
          muted: activeRouteTarget && !route.active,
        }"
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
        data-board-focus
        :data-board-file="piece.file"
        :data-board-rank="piece.rank"
        :tabindex="activeFocusTarget === `piece-${piece.renderId}` ? 0 : -1"
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
        @focus="focusedBoardTarget = `piece-${piece.renderId}`"
        @keydown="handleBoardArrow"
        @click.stop="handlePieceClick(piece)"
      ><span class="board-piece-glyph">{{ piece.name }}</span></button>
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
      data-board-focus
      :data-board-file="hint.file"
      :data-board-rank="hint.rank"
      :tabindex="activeFocusTarget === `hint-${hint.file}-${hint.rank}` ? 0 : -1"
      :class="{ capture: isOccupied(hint.file, hint.rank) }"
      :style="screenPosition(hint.file, hint.rank, match.flipped)"
      :aria-label="`${isOccupied(hint.file, hint.rank) ? '吃子并移动' : '移动'}到 ${hint.file},${hint.rank}`"
      @focus="focusedBoardTarget = `hint-${hint.file}-${hint.rank}`"
      @focusin="activateRoute(hint.file, hint.rank)"
      @focusout="clearActiveRoute(hint.file, hint.rank)"
      @pointerenter="activateRoute(hint.file, hint.rank)"
      @pointerleave="clearActiveRoute(hint.file, hint.rank)"
      @keydown="handleBoardArrow"
      @click.stop="move(hint.file, hint.rank)"
    />
    <span id="board-keyboard-help" class="sr-only">按方向键在棋子和合法落点之间移动，按回车或空格选择。</span>
    <span class="sr-only" aria-live="polite">
      {{ match.legalMovesLoading ? '正在获取合法落点' : match.selectedPos ? `找到 ${match.hints.length} 个合法落点` : '' }}
    </span>
    <div
      v-if="match.inCheck && !match.isFinished && !isAnimating && !captureMarker"
      class="board-check-callout board-state-ribbon"
      role="status"
      aria-live="assertive"
    >
      <span class="board-state-seal">将</span>
      <span>
        <strong>{{ match.sideToMove === match.playerColor ? '你被将军' : '将军' }}</strong>
        <small>{{ match.sideToMove === match.playerColor ? '请选择合法着法应将' : '对方正在应将' }}</small>
      </span>
    </div>
    <div
      v-if="captureMarker && capturePresentation && !match.isFinished"
      :key="captureMarker.key"
      class="capture-event-layer"
      :style="screenPosition(captureMarker.file, captureMarker.rank, match.flipped)"
      role="status"
      aria-live="assertive"
      aria-atomic="true"
      :aria-label="capturePresentation.announcement"
    >
      <span class="capture-impact-marker" aria-hidden="true" />
      <span class="capture-event-callout" :class="capturePresentation.placement" aria-hidden="true">
        <b>吃</b><strong>{{ capturePresentation.label }}</strong>
      </span>
    </div>
    <section
      v-if="checkmatePresentation && !isAnimating && !captureMarker"
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
        <div class="board-finish-actions">
          <button class="primary-button" @click.stop="emit('restart')"><AppIcon name="refresh" />再来一局</button>
          <button class="secondary-button" @click.stop="emit('review')"><AppIcon name="chart" />查看复盘</button>
        </div>
      </div>
    </section>
  </div>
  <slot />
  <div class="board-toolbar" aria-label="棋盘工具">
    <button :disabled="!match.allowUndo || match.isFinished || isAnimating" @click="$emit('undo')"><AppIcon name="undo" /><span>悔棋</span></button>
    <button :disabled="isAnimating" @click="flip"><AppIcon name="flip" /><span>翻转</span></button>
    <button
      :class="{ 'sound-enabled': match.soundEnabled }"
      :aria-pressed="match.soundEnabled"
      @click="toggleSound"
    ><AppIcon :name="match.soundEnabled ? 'volume' : 'volume-off'" /><span>音效</span></button>
    <button class="danger-ghost" :disabled="match.isFinished" @click="$emit('resign')"><AppIcon name="flag" /><span>认输</span></button>
  </div>
</template>
