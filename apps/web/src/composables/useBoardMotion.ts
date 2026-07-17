import { onScopeDispose, ref, watch, type Ref } from 'vue'
import type { MoveRecord } from '@/api/contracts'
import type { BoardPiece, BoardSquare } from '@/types/xiangqi'
import { getPiecesFromFEN } from '@/utils/board'

export type PieceMotion = 'idle' | 'lifting' | 'moving' | 'settling' | 'captured' | 'restored'
export type BoardMotionPhase = 'idle' | 'lifting' | 'moving' | 'capturing' | 'settling'

export interface MotionBoardPiece extends BoardPiece {
  renderId: string
  motion: PieceMotion
  motionDurationMs: number
}

export interface BoardArrivalMarker extends BoardSquare {
  key: number
}

export interface BoardCaptureMarker extends BoardSquare {
  key: number
  captured: string
  actor: MoveRecord['actor']
}

interface BoardMotionOptions {
  fen: Ref<string>
  moves: Ref<MoveRecord[]>
  matchId: Ref<string | null>
}

interface QueuedMove {
  direction: 'forward' | 'reverse'
  record: MoveRecord
}

export interface BoardMoveSquares {
  from: BoardSquare
  to: BoardSquare
}

const MOVE_NEAR_MS = 240
const MOVE_FAR_MS = 420
const AI_MOVE_EXTRA_MS = 20
const UNDO_NEAR_MS = 220
const UNDO_FAR_MS = 340
const PIECE_LIFT_MS = 55
const PIECE_SETTLE_MS = 70
const ARRIVAL_MARKER_MS = 360
const CAPTURE_MARKER_MS = 680

function parseMove(move: string): BoardMoveSquares | null {
  if (!/^[a-i][0-9][a-i][0-9]$/.test(move)) return null

  return {
    from: {
      file: move.charCodeAt(0) - 97,
      rank: 9 - Number.parseInt(move[1]!, 10),
    },
    to: {
      file: move.charCodeAt(2) - 97,
      rank: 9 - Number.parseInt(move[3]!, 10),
    },
  }
}

function sameSquare(piece: BoardSquare, square: BoardSquare) {
  return piece.file === square.file && piece.rank === square.rank
}

function samePiece(left: BoardPiece, right: BoardPiece) {
  return left.color === right.color && left.name === right.name
}

function sameMove(left: MoveRecord, right: MoveRecord) {
  return left.ply === right.ply
    && left.move === right.move
    && left.fenBefore === right.fenBefore
    && left.fenAfter === right.fenAfter
}

function isMovePrefix(prefix: MoveRecord[], moves: MoveRecord[]) {
  return prefix.length <= moves.length && prefix.every((move, index) => sameMove(move, moves[index]!))
}

function dedupeMoves(moves: MoveRecord[]) {
  const byPly = new Map<number, MoveRecord>()
  for (const move of moves) byPly.set(move.ply, move)
  return [...byPly.values()].sort((left, right) => left.ply - right.ply)
}

function targetOf(move: MoveRecord | undefined) {
  if (!move) return null
  return parseMove(move.move)?.to ?? null
}

function boardSignature(pieces: BoardPiece[]) {
  return pieces
    .map((piece) => `${piece.file},${piece.rank},${piece.color},${piece.name}`)
    .sort()
    .join('|')
}

function moveDuration(record: MoveRecord, parsed: BoardMoveSquares, reverse: boolean) {
  const fileDistance = parsed.to.file - parsed.from.file
  const rankDistance = parsed.to.rank - parsed.from.rank
  const distance = Math.min(8, Math.hypot(fileDistance, rankDistance))
  const progress = Math.max(0, (distance - 1) / 7)

  if (reverse) {
    return Math.round(UNDO_NEAR_MS + (UNDO_FAR_MS - UNDO_NEAR_MS) * progress)
  }

  const actorOffset = record.actor === 'ai' ? AI_MOVE_EXTRA_MS : 0
  return Math.min(
    MOVE_FAR_MS,
    Math.round(MOVE_NEAR_MS + (MOVE_FAR_MS - MOVE_NEAR_MS) * progress + actorOffset),
  )
}

export function useBoardMotion(options: BoardMotionOptions) {
  let renderSequence = 0
  let markerSequence = 0
  let captureMarkerSequence = 0
  let generation = 0
  let processing = false
  let observedMatchId = options.matchId.value
  let observedMoves = dedupeMoves(options.moves.value)
  let latestFen = options.fen.value
  let latestMoves = observedMoves
  let markerTimer: ReturnType<typeof setTimeout> | null = null
  let captureMarkerTimer: ReturnType<typeof setTimeout> | null = null

  const queue: QueuedMove[] = []
  const pendingWaits = new Map<ReturnType<typeof setTimeout>, () => void>()
  const pieces = ref<MotionBoardPiece[]>([])
  const isAnimating = ref(false)
  const lastSquare = ref<BoardSquare | null>(targetOf(latestMoves.at(-1)))
  const lastMoveSquares = ref<BoardMoveSquares | null>(
    latestMoves.at(-1) ? parseMove(latestMoves.at(-1)!.move) : null,
  )
  const arrivalMarker = ref<BoardArrivalMarker | null>(null)
  const captureMarker = ref<BoardCaptureMarker | null>(null)
  const motionPhase = ref<BoardMotionPhase>('idle')

  const reducedMotion = ref(false)
  const motionQuery = typeof window !== 'undefined' && typeof window.matchMedia === 'function'
    ? window.matchMedia('(prefers-reduced-motion: reduce)')
    : null

  if (motionQuery) {
    reducedMotion.value = motionQuery.matches
    motionQuery.addEventListener?.('change', handleMotionPreference)
  }

  function handleMotionPreference(event: MediaQueryListEvent) {
    reducedMotion.value = event.matches
  }

  function freshPiece(piece: BoardPiece, motion: PieceMotion = 'idle'): MotionBoardPiece {
    renderSequence += 1
    return {
      ...piece,
      renderId: `board-piece-${renderSequence}`,
      motion,
      motionDurationMs: 0,
    }
  }

  function activePieces() {
    return pieces.value.filter((piece) => piece.motion !== 'captured')
  }

  function matchesFen(fen: string) {
    return boardSignature(activePieces()) === boardSignature(getPiecesFromFEN(fen))
  }

  function syncToFen(fen: string) {
    const current = activePieces()
    const target = getPiecesFromFEN(fen)

    pieces.value = target.map((piece) => {
      const existing = current.find((candidate) =>
        sameSquare(candidate, piece) && samePiece(candidate, piece),
      )
      return existing
        ? { ...existing, ...piece, motion: 'idle', motionDurationMs: 0 }
        : freshPiece(piece)
    })
  }

  function settlePieces(nextPieces: MotionBoardPiece[]) {
    pieces.value = nextPieces
      .filter((piece) => piece.motion !== 'captured')
      .map((piece) => ({ ...piece, motion: 'idle', motionDurationMs: 0 }))
  }

  function liftMovingPiece(record: MoveRecord, parsed: BoardMoveSquares, reverse: boolean) {
    const sourceFen = reverse ? record.fenAfter : record.fenBefore
    if (!matchesFen(sourceFen)) syncToFen(sourceFen)
    const sourceSquare = reverse ? parsed.to : parsed.from
    const movingPiece = activePieces().find((piece) => sameSquare(piece, sourceSquare))
    if (!movingPiece) return null

    pieces.value = pieces.value.map((piece) => piece.renderId === movingPiece.renderId
      ? { ...piece, motion: 'lifting' as const, motionDurationMs: 0 }
      : piece)
    return movingPiece.renderId
  }

  function wait(durationMs: number) {
    if (durationMs <= 0) return Promise.resolve()

    return new Promise<void>((resolve) => {
      const timer = setTimeout(() => {
        pendingWaits.delete(timer)
        resolve()
      }, durationMs)
      pendingWaits.set(timer, resolve)
    })
  }

  function flashArrival(square: BoardSquare) {
    if (markerTimer) clearTimeout(markerTimer)
    if (reducedMotion.value) {
      arrivalMarker.value = null
      return
    }

    markerSequence += 1
    const marker = { ...square, key: markerSequence }
    arrivalMarker.value = marker
    markerTimer = setTimeout(() => {
      if (arrivalMarker.value?.key === marker.key) arrivalMarker.value = null
      markerTimer = null
    }, ARRIVAL_MARKER_MS)
  }

  function flashCapture(record: MoveRecord, square: BoardSquare) {
    if (!record.captured) return
    if (captureMarkerTimer) clearTimeout(captureMarkerTimer)

    captureMarkerSequence += 1
    const marker = {
      ...square,
      key: captureMarkerSequence,
      captured: record.captured,
      actor: record.actor,
    }
    captureMarker.value = marker
    captureMarkerTimer = setTimeout(() => {
      if (captureMarker.value?.key === marker.key) captureMarker.value = null
      captureMarkerTimer = null
    }, CAPTURE_MARKER_MS)
  }

  function clearCaptureMarker() {
    if (captureMarkerTimer) clearTimeout(captureMarkerTimer)
    captureMarkerTimer = null
    captureMarker.value = null
  }

  function buildForwardPieces(record: MoveRecord, parsed: BoardMoveSquares, durationMs: number) {
    if (!matchesFen(record.fenBefore)) syncToFen(record.fenBefore)

    const current = activePieces()
    const movingPiece = current.find((piece) => sameSquare(piece, parsed.from))
    if (!movingPiece) return null

    const capturedPiece = current.find((piece) =>
      piece.renderId !== movingPiece.renderId && sameSquare(piece, parsed.to),
    )
    const target = getPiecesFromFEN(record.fenAfter)

    const next = target.map((piece) => {
      if (sameSquare(piece, parsed.to) && samePiece(piece, movingPiece)) {
        return {
          ...movingPiece,
          ...piece,
          motion: 'moving' as const,
          motionDurationMs: durationMs,
        }
      }

      const existing = current.find((candidate) =>
        candidate.renderId !== movingPiece.renderId
        && candidate.renderId !== capturedPiece?.renderId
        && sameSquare(candidate, piece)
        && samePiece(candidate, piece),
      )
      return existing
        ? { ...existing, ...piece, motion: 'idle' as const, motionDurationMs: 0 }
        : freshPiece(piece)
    })

    if (capturedPiece) {
      next.push({
        ...capturedPiece,
        motion: 'captured',
        motionDurationMs: durationMs,
      })
    }

    return next
  }

  function buildReversePieces(record: MoveRecord, parsed: BoardMoveSquares, durationMs: number) {
    if (!matchesFen(record.fenAfter)) syncToFen(record.fenAfter)

    const current = activePieces()
    const movingPiece = current.find((piece) => sameSquare(piece, parsed.to))
    if (!movingPiece) return null

    const target = getPiecesFromFEN(record.fenBefore)
    return target.map((piece) => {
      if (sameSquare(piece, parsed.from) && samePiece(piece, movingPiece)) {
        return {
          ...movingPiece,
          ...piece,
          motion: 'moving' as const,
          motionDurationMs: durationMs,
        }
      }

      const existing = current.find((candidate) =>
        candidate.renderId !== movingPiece.renderId
        && sameSquare(candidate, piece)
        && samePiece(candidate, piece),
      )
      return existing
        ? { ...existing, ...piece, motion: 'idle' as const, motionDurationMs: 0 }
        : freshPiece(piece, 'restored')
    })
  }

  async function animateMove(queuedMove: QueuedMove, activeGeneration: number) {
    clearCaptureMarker()
    const parsed = parseMove(queuedMove.record.move)
    if (!parsed) {
      syncToFen(
        queuedMove.direction === 'forward'
          ? queuedMove.record.fenAfter
          : queuedMove.record.fenBefore,
      )
      return
    }

    const reverse = queuedMove.direction === 'reverse'
    const durationMs = reducedMotion.value ? 0 : moveDuration(queuedMove.record, parsed, reverse)
    const liftedPieceId = liftMovingPiece(queuedMove.record, parsed, reverse)
    if (!liftedPieceId) {
      syncToFen(reverse ? queuedMove.record.fenBefore : queuedMove.record.fenAfter)
      return
    }

    if (durationMs > 0) {
      motionPhase.value = 'lifting'
      await wait(PIECE_LIFT_MS)
      if (activeGeneration !== generation) return
    }

    const nextPieces = reverse
      ? buildReversePieces(queuedMove.record, parsed, durationMs)
      : buildForwardPieces(queuedMove.record, parsed, durationMs)

    if (!nextPieces) {
      syncToFen(reverse ? queuedMove.record.fenBefore : queuedMove.record.fenAfter)
      return
    }

    motionPhase.value = !reverse && queuedMove.record.captured ? 'capturing' : 'moving'
    pieces.value = nextPieces
    await wait(durationMs)
    if (activeGeneration !== generation) return

    const destination = reverse ? parsed.from : parsed.to
    const landingPieces = nextPieces
      .filter((piece) => piece.motion !== 'captured')
      .map((piece) => piece.renderId === liftedPieceId
        ? { ...piece, motion: 'settling' as const, motionDurationMs: 0 }
        : { ...piece, motion: 'idle' as const, motionDurationMs: 0 })
    pieces.value = landingPieces
    motionPhase.value = 'settling'
    lastSquare.value = destination
    flashArrival(destination)
    if (!reverse) flashCapture(queuedMove.record, destination)
    await wait(reducedMotion.value ? 0 : PIECE_SETTLE_MS)
    if (activeGeneration !== generation) return
    settlePieces(landingPieces)
  }

  async function processQueue() {
    if (processing) return

    processing = true
    isAnimating.value = true
    const activeGeneration = generation

    while (queue.length > 0 && activeGeneration === generation) {
      const queuedMove = queue.shift()
      if (queuedMove) await animateMove(queuedMove, activeGeneration)
    }

    if (activeGeneration === generation) {
      if (!matchesFen(latestFen)) syncToFen(latestFen)
      lastSquare.value = targetOf(latestMoves.at(-1))
      lastMoveSquares.value = latestMoves.at(-1)
        ? parseMove(latestMoves.at(-1)!.move)
        : null
      processing = false
      isAnimating.value = false
      motionPhase.value = 'idle'
    }
  }

  function enqueue(queuedMoves: QueuedMove[]) {
    queue.push(...queuedMoves)
    void processQueue()
  }

  function cancelAnimations() {
    generation += 1
    queue.length = 0
    for (const [timer, resolve] of pendingWaits) {
      clearTimeout(timer)
      resolve()
    }
    pendingWaits.clear()
    if (markerTimer) clearTimeout(markerTimer)
    markerTimer = null
    arrivalMarker.value = null
    clearCaptureMarker()
    processing = false
    isAnimating.value = false
    motionPhase.value = 'idle'
  }

  function resetBoard(fen: string, moves: MoveRecord[]) {
    cancelAnimations()
    syncToFen(fen)
    lastSquare.value = targetOf(moves.at(-1))
    lastMoveSquares.value = moves.at(-1) ? parseMove(moves.at(-1)!.move) : null
  }

  syncToFen(options.fen.value)

  watch(
    [options.matchId, options.fen, options.moves],
    ([nextMatchId, nextFen, rawNextMoves]) => {
      const nextMoves = dedupeMoves(rawNextMoves)
      latestFen = nextFen
      latestMoves = nextMoves

      if (nextMatchId !== observedMatchId) {
        observedMatchId = nextMatchId
        observedMoves = nextMoves
        resetBoard(nextFen, nextMoves)
        return
      }

      if (isMovePrefix(observedMoves, nextMoves) && nextMoves.length > observedMoves.length) {
        const added = nextMoves.slice(observedMoves.length)
        observedMoves = nextMoves
        enqueue(added.map((record) => ({ direction: 'forward', record })))
        return
      }

      if (isMovePrefix(nextMoves, observedMoves) && nextMoves.length < observedMoves.length) {
        const removed = observedMoves.slice(nextMoves.length).reverse()
        observedMoves = nextMoves
        enqueue(removed.map((record) => ({ direction: 'reverse', record })))
        return
      }

      const historyUnchanged = observedMoves.length === nextMoves.length
        && observedMoves.every((move, index) => sameMove(move, nextMoves[index]!))
      observedMoves = nextMoves

      if (!historyUnchanged && !processing) {
        resetBoard(nextFen, nextMoves)
      } else if (historyUnchanged && !processing && !matchesFen(nextFen)) {
        syncToFen(nextFen)
      }
    },
  )

  onScopeDispose(() => {
    cancelAnimations()
    motionQuery?.removeEventListener?.('change', handleMotionPreference)
  })

  return {
    pieces,
    isAnimating,
    lastSquare,
    lastMoveSquares,
    arrivalMarker,
    captureMarker,
    motionPhase,
    reducedMotion,
  }
}
