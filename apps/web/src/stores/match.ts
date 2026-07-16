import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useRouter } from 'vue-router'
import { ApiError } from '@/api/client'
import {
  createMatch as requestCreateMatch,
  getLegalMoves,
  getMatch,
  offerMatchDraw,
  resignMatch,
  submitMatchMove,
  undoMatch,
} from '@/api/matches'
import { connectMatchStream } from '@/api/stream'
import {
  fromICCSSquare,
  getPiecesFromFEN,
  toICCSCode,
  toICCSSquare,
} from '@/utils/board'
import { getSideToMove } from '@/utils/fen'
import { useUiStore } from './ui'
import type { AiMode, BoardPiece, BoardSquare, Color } from '@/types/xiangqi'
import type {
  MatchEvent,
  MatchSnapshot,
  MatchStatus,
  MatchOutcome,
  MoveRecord,
} from '@/api/contracts'

export const useMatchStore = defineStore('match', () => {
  // ── 对局状态（来自后端） ──
  const matchId = ref<string | null>(null)
  const version = ref(0)
  const status = ref<MatchStatus>('active_player_turn')
  const playerColor = ref<Color>('red')
  const sideToMove = ref<Color>('red')
  const difficulty = ref(0)
  const aiMode = ref<AiMode>('standard')
  const engine = ref('')
  const allowUndo = ref(true)
  const initialFen = ref('')
  const fen = ref('')
  const moves = ref<MoveRecord[]>([])
  const outcome = ref<MatchOutcome>('ongoing')
  const drawOffered = ref(false)

  // ── UI 交互状态 ──
  const flipped = ref(false)
  const soundEnabled = ref(true)
  const selectedPos = ref<{ file: number; rank: number } | null>(null)
  const hints = ref<BoardSquare[]>([])
  const legalMovesLoading = ref(false)
  const thinking = ref(false)
  const rejectedMove = ref<{ id: number; file: number; rank: number } | null>(null)

  // ── 派生状态 ──
  const pieces = computed<BoardPiece[]>(() => getPiecesFromFEN(fen.value))

  const myTurn = computed(() =>
    status.value === 'active_player_turn' && sideToMove.value === playerColor.value,
  )

  const isFinished = computed(() =>
    status.value === 'finished' || status.value === 'aborted',
  )

  const statusLabel = computed(() => {
    if (status.value === 'finished') return '对局已结束'
    if (status.value === 'active_player_turn') return '你的回合'
    if (status.value === 'active_ai_thinking') return 'AI 思考中'
    if (status.value === 'recoverable_error') return '引擎恢复中'
    return '对局已中止'
  })

  const lastMove = computed<MoveRecord | null>(() =>
    moves.value.length > 0 ? moves.value[moves.value.length - 1]! : null,
  )

  // ── 内部状态 ──
  let streamHandle: ReturnType<typeof connectMatchStream> | null = null
  let rejectedMoveSequence = 0
  let legalMoveRequestSequence = 0
  let resyncPromise: Promise<void> | null = null
  const legalMoveCache = new Map<string, BoardSquare[]>()
  const router = useRouter()

  function legalMoveCacheKey(
    id: string,
    matchVersion: number,
    file: number,
    rank: number,
  ) {
    return `${id}:${matchVersion}:${toICCSSquare(file, rank)}`
  }

  function invalidateLegalMoveCache(clearSelected = true) {
    legalMoveCache.clear()
    legalMoveRequestSequence += 1
    legalMovesLoading.value = false
    hints.value = []
    if (clearSelected) selectedPos.value = null
  }

  function upsertMove(move: MoveRecord) {
    const existingIndex = moves.value.findIndex((candidate) => candidate.ply === move.ply)
    if (existingIndex >= 0) {
      const existing = moves.value[existingIndex]!
      if (existing.move === move.move && existing.fenAfter === move.fenAfter) return false
      const nextMoves = [...moves.value]
      nextMoves[existingIndex] = move
      moves.value = nextMoves
      return true
    }

    moves.value = [...moves.value, move].sort((left, right) => left.ply - right.ply)
    return true
  }

  // ── 从 MatchSnapshot 填充状态 ──
  function applySnapshot(snapshot: MatchSnapshot) {
    if (
      snapshot.id === matchId.value
      && snapshot.version < version.value
    ) return

    if (snapshot.id !== matchId.value || snapshot.version !== version.value) {
      invalidateLegalMoveCache()
    }
    matchId.value = snapshot.id
    version.value = snapshot.version
    status.value = snapshot.status
    playerColor.value = snapshot.playerColor as Color
    sideToMove.value = snapshot.sideToMove as Color
    difficulty.value = snapshot.difficulty
    aiMode.value = snapshot.aiMode ?? 'standard'
    engine.value = snapshot.engine
    allowUndo.value = snapshot.allowUndo
    initialFen.value = snapshot.initialFen
    fen.value = snapshot.fen
    moves.value = snapshot.moves ?? []
    outcome.value = snapshot.outcome
    drawOffered.value = snapshot.drawOffered ?? false
    thinking.value = snapshot.status === 'active_ai_thinking'
  }

  async function synchronizeMatch(id: string) {
    if (resyncPromise) return resyncPromise

    resyncPromise = getMatch(id)
      .then((snapshot) => {
        if (matchId.value === null || snapshot.id === matchId.value) {
          applySnapshot(snapshot)
        }
      })
      .finally(() => {
        resyncPromise = null
      })
    return resyncPromise
  }

  // ── 事件处理 ──
  function handleEvent(event: MatchEvent) {
    if (matchId.value && event.matchId !== matchId.value) return
    // 忽略旧版本事件
    if (event.matchVersion < version.value) return
    if (event.matchVersion > version.value + 1) {
      void synchronizeMatch(event.matchId).catch(() => {
        const ui = useUiStore()
        ui.showToast('实时状态有缺口，请刷新后继续')
      })
      return
    }

    if (event.matchVersion !== version.value) {
      invalidateLegalMoveCache()
    }
    version.value = event.matchVersion

    switch (event.type) {
      case 'match.snapshot':
      case 'match.finished':
        applySnapshot(event.payload as MatchSnapshot)
        break

      case 'match.move_accepted': {
        const moveRecord = event.payload as MoveRecord
        upsertMove(moveRecord)
        status.value = 'active_ai_thinking'
        thinking.value = true
        if (moveRecord.fenAfter) {
          fen.value = moveRecord.fenAfter
          sideToMove.value = getSideToMove(moveRecord.fenAfter)
        }
        break
      }

      case 'match.ai_thinking':
        thinking.value = true
        status.value = 'active_ai_thinking'
        break

      case 'match.ai_move_applied': {
        const payload = event.payload as {
          move: MoveRecord
          depth: number
          nodes: number
          stoppedReason: string
        }
        upsertMove(payload.move)
        if (payload.move.fenAfter) {
          fen.value = payload.move.fenAfter
          sideToMove.value = getSideToMove(payload.move.fenAfter)
        }
        status.value = 'active_player_turn'
        thinking.value = false
        break
      }

      case 'match.undo_applied':
        applySnapshot(event.payload as MatchSnapshot)
        break

      case 'match.engine_degraded': {
        const reason = (event.payload as { reason: string }).reason
        const ui = useUiStore()
        ui.showToast(`引擎状态异常：${reason}`)
        break
      }

      case 'match.draw_declined': {
        const ui = useUiStore()
        ui.showToast('对方拒绝了和棋请求')
        drawOffered.value = false
        break
      }
    }
  }

  // ── 建立 WebSocket 连接 ──
  function connectStream(id: string) {
    disconnectStream()
    streamHandle = connectMatchStream(id)
    streamHandle.onEvent(handleEvent)
  }

  function disconnectStream() {
    if (streamHandle) {
      streamHandle.close()
      streamHandle = null
    }
  }

  // ── 创建对局 ──
  async function createMatch(
    playerColorParam: Color,
    difficultyParam: number,
    aiModeParam: AiMode = 'standard',
    allowUndoParam = true,
  ) {
    const snapshot = await requestCreateMatch({
      playerColor: playerColorParam,
      difficulty: difficultyParam,
      aiMode: aiModeParam,
      allowUndo: allowUndoParam,
    })
    applySnapshot(snapshot)
    connectStream(snapshot.id)
    await router.push(`/match/${snapshot.id}`)
    return snapshot
  }

  // ── 加载已有对局 ──
  async function loadMatch(id: string) {
    const snapshot = await getMatch(id)
    applySnapshot(snapshot)
    connectStream(id)
  }

  // ── 选取棋子 ──
  async function selectPieceAt(file: number, rank: number) {
    const piece = pieces.value.find((p) => p.file === file && p.rank === rank)
    if (!piece || piece.color !== playerColor.value || !myTurn.value || !matchId.value) return

    selectedPos.value = { file, rank }
    hints.value = []
    const requestID = ++legalMoveRequestSequence
    const requestMatchID = matchId.value
    const requestVersion = version.value
    const cacheKey = legalMoveCacheKey(requestMatchID, requestVersion, file, rank)
    const cached = legalMoveCache.get(cacheKey)
    if (cached) {
      hints.value = [...cached]
      return
    }

    legalMovesLoading.value = true
    try {
      const response = await getLegalMoves(requestMatchID, toICCSSquare(file, rank))
      if (
        requestID !== legalMoveRequestSequence
        || matchId.value !== requestMatchID
        || version.value !== requestVersion
        || selectedPos.value?.file !== file
        || selectedPos.value?.rank !== rank
      ) return

      if (response.matchVersion !== requestVersion) {
        await synchronizeMatch(requestMatchID)
        return
      }
      const authoritativeHints = response.moves
        .map((move) => fromICCSSquare(move.to))
        .filter((square): square is BoardSquare => square !== null)
      legalMoveCache.set(cacheKey, authoritativeHints)
      hints.value = [...authoritativeHints]
    } catch {
      if (requestID === legalMoveRequestSequence) {
        const ui = useUiStore()
        ui.showToast('暂时无法获取合法落点，请稍后重试')
      }
    } finally {
      if (requestID === legalMoveRequestSequence) {
        legalMovesLoading.value = false
      }
    }
  }

  function clearSelection() {
    legalMoveRequestSequence += 1
    legalMovesLoading.value = false
    selectedPos.value = null
    hints.value = []
  }

  function restoreSelection(file: number, rank: number) {
    if (!matchId.value) return
    selectedPos.value = { file, rank }
    const cached = legalMoveCache.get(
      legalMoveCacheKey(matchId.value, version.value, file, rank),
    )
    hints.value = cached ? [...cached] : []
  }

  // ── 提交着法 ──
  async function submitMove(fromFile: number, fromRank: number, toFile: number, toRank: number) {
    if (!matchId.value) return
    const submittedMatchID = matchId.value
    const iccs = toICCSCode(fromFile, fromRank, toFile, toRank)

    // 乐观更新：切换为 AI 思考状态
    const prevStatus = status.value
    status.value = 'active_ai_thinking'
    thinking.value = true
    clearSelection()

    try {
      const snapshot = await submitMatchMove(submittedMatchID, {
        move: iccs,
        expectedMatchVersion: version.value,
      })
      // 用服务端快照同步状态
      applySnapshot(snapshot)
    } catch (error) {
      // 请求失败，恢复状态
      status.value = prevStatus
      thinking.value = false
      rejectedMoveSequence += 1
      rejectedMove.value = { id: rejectedMoveSequence, file: fromFile, rank: fromRank }
      restoreSelection(fromFile, fromRank)
      const ui = useUiStore()
      if (error instanceof ApiError) {
        switch (error.body.code) {
          case 'ILLEGAL_MOVE':
            ui.showToast('该着法不合法')
            return
          case 'MATCH_VERSION_CONFLICT':
            ui.showToast('局面已更新，正在同步')
            try {
              await synchronizeMatch(submittedMatchID)
            } catch {
              ui.showToast('同步最新局面失败，请刷新页面')
            }
            return
          case 'MATCH_NOT_FOUND':
          case 'NOT_FOUND':
            ui.showToast('对局不存在或已失效')
            reset()
            await router.push('/history')
            return
          default:
            ui.showToast(error.body.message || '着法提交失败，请重试')
            return
        }
      }
      ui.showToast('网络连接异常，请重试')
    }
  }

  // ── 悔棋 ──
  async function undo() {
    if (!matchId.value || !allowUndo.value) return false
    try {
      const snapshot = await undoMatch(matchId.value, {
        expectedMatchVersion: version.value,
      })
      applySnapshot(snapshot)
      clearSelection()
      const ui = useUiStore()
      ui.showToast('已撤销落子')
      return true
    } catch {
      const ui = useUiStore()
      ui.showToast('悔棋失败')
      return false
    }
  }

  // ── 认输 ──
  async function resign() {
    if (!matchId.value) return false
    try {
      const snapshot = await resignMatch(matchId.value, {
        expectedMatchVersion: version.value,
      })
      applySnapshot(snapshot)
      clearSelection()
      return true
    } catch {
      const ui = useUiStore()
      ui.showToast('认输请求失败')
      return false
    }
  }

  // ── 请求和棋 ──
  async function offerDraw() {
    if (!matchId.value) return false
    try {
      const result = await offerMatchDraw(matchId.value, {
        expectedMatchVersion: version.value,
      })
      applySnapshot(result.match)
      if (result.accepted) {
        const ui = useUiStore()
        ui.showToast('和棋已接受')
      } else {
        drawOffered.value = true
        const ui = useUiStore()
        ui.showToast('已提出和棋')
      }
      return result.accepted
    } catch {
      const ui = useUiStore()
      ui.showToast('和棋请求失败')
      return false
    }
  }

  // ── 重置（回到初始局面） ──
  function reset() {
    disconnectStream()
    invalidateLegalMoveCache()
    matchId.value = null
    version.value = 0
    status.value = 'active_player_turn'
    playerColor.value = 'red'
    sideToMove.value = 'red'
    difficulty.value = 0
    aiMode.value = 'standard'
    engine.value = ''
    allowUndo.value = true
    initialFen.value = ''
    fen.value = ''
    moves.value = []
    outcome.value = 'ongoing'
    drawOffered.value = false
    thinking.value = false
    rejectedMove.value = null
  }

  // ── 清理 ──
  function dispose() {
    disconnectStream()
    reset()
  }

  return {
    // 状态
    matchId, version, status, playerColor, sideToMove, difficulty,
    engine, aiMode, allowUndo, initialFen, fen, moves, outcome, drawOffered,
    // UI
    flipped, soundEnabled, selectedPos, hints, legalMovesLoading, thinking, rejectedMove,
    // 派生
    pieces, myTurn, isFinished, statusLabel, lastMove,
    // 方法
    createMatch, loadMatch, selectPieceAt, clearSelection, submitMove,
    undo, resign, offerDraw, reset, dispose,
  }
})
