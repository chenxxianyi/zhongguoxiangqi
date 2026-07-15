import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { useRouter } from 'vue-router'
import { apiRequest } from '@/api/client'
import { connectMatchStream } from '@/api/stream'
import { getPiecesFromFEN, toICCSCode, candidateMoves } from '@/utils/board'
import { initialFEN } from '@/utils/fen'
import { useUiStore } from './ui'
import type { BoardPiece, BoardSquare, Color } from '@/types/xiangqi'
import type { MatchSnapshot, MatchStatus, MatchOutcome, MoveRecord } from '@/api/contracts'

export const useMatchStore = defineStore('match', () => {
  // ── 对局状态（来自后端） ──
  const matchId = ref<string | null>(null)
  const version = ref(0)
  const status = ref<MatchStatus>('active_player_turn')
  const playerColor = ref<Color>('red')
  const sideToMove = ref<Color>('red')
  const difficulty = ref(0)
  const engine = ref('')
  const allowUndo = ref(true)
  const fen = ref(initialFEN)
  const moves = ref<MoveRecord[]>([])
  const outcome = ref<MatchOutcome>('ongoing')
  const drawOffered = ref(false)

  // ── UI 交互状态 ──
  const flipped = ref(false)
  const soundEnabled = ref(true)
  const selectedPos = ref<{ file: number; rank: number } | null>(null)
  const hints = ref<BoardSquare[]>([])
  const thinking = ref(false)

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
  const router = useRouter()

  // ── 从 MatchSnapshot 填充状态 ──
  function applySnapshot(snapshot: MatchSnapshot) {
    matchId.value = snapshot.id
    version.value = snapshot.version
    status.value = snapshot.status
    playerColor.value = snapshot.playerColor as Color
    sideToMove.value = snapshot.sideToMove as Color
    difficulty.value = snapshot.difficulty
    engine.value = snapshot.engine
    allowUndo.value = snapshot.allowUndo
    fen.value = snapshot.fen
    moves.value = snapshot.moves ?? []
    outcome.value = snapshot.outcome
    drawOffered.value = snapshot.drawOffered ?? false
    thinking.value = snapshot.status === 'active_ai_thinking'
  }

  // ── 事件处理 ──
  function handleEvent(event: { type: string; payload: unknown; matchVersion: number }) {
    // 忽略旧版本事件
    if (event.matchVersion < version.value) return

    version.value = event.matchVersion

    switch (event.type) {
      case 'match.snapshot':
      case 'match.finished':
        applySnapshot(event.payload as MatchSnapshot)
        break

      case 'match.move_accepted': {
        const moveRecord = event.payload as MoveRecord
        moves.value = [...moves.value, moveRecord]
        // 切换行棋方
        sideToMove.value = sideToMove.value === 'red' ? 'black' : 'red'
        status.value = 'active_ai_thinking'
        thinking.value = true
        // 更新 FEN
        if (moveRecord.fenAfter) {
          fen.value = moveRecord.fenAfter
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
        moves.value = [...moves.value, payload.move]
        if (payload.move.fenAfter) {
          fen.value = payload.move.fenAfter
        }
        sideToMove.value = sideToMove.value === 'red' ? 'black' : 'red'
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
  async function createMatch(playerColorParam: Color, difficultyParam: number, allowUndoParam = true) {
    const snapshot = await apiRequest<MatchSnapshot>('/matches', {
      method: 'POST',
      body: JSON.stringify({
        playerColor: playerColorParam,
        difficulty: difficultyParam,
        allowUndo: allowUndoParam,
      }),
    })
    applySnapshot(snapshot)
    connectStream(snapshot.id)
    await router.push(`/match/${snapshot.id}`)
    return snapshot
  }

  // ── 加载已有对局 ──
  async function loadMatch(id: string) {
    const snapshot = await apiRequest<MatchSnapshot>(`/matches/${id}`)
    applySnapshot(snapshot)
    connectStream(id)
  }

  // ── 选取棋子 ──
  function selectPieceAt(file: number, rank: number) {
    const piece = pieces.value.find((p) => p.file === file && p.rank === rank)
    if (!piece) return
    selectedPos.value = { file, rank }
    hints.value = candidateMoves(piece)
  }

  function clearSelection() {
    selectedPos.value = null
    hints.value = []
  }

  // ── 提交着法 ──
  async function submitMove(fromFile: number, fromRank: number, toFile: number, toRank: number) {
    if (!matchId.value) return
    const iccs = toICCSCode(fromFile, fromRank, toFile, toRank)

    // 乐观更新：切换为 AI 思考状态
    const prevStatus = status.value
    status.value = 'active_ai_thinking'
    thinking.value = true
    clearSelection()

    try {
      const snapshot = await apiRequest<MatchSnapshot>(`/matches/${matchId.value}/moves`, {
        method: 'POST',
        body: JSON.stringify({
          move: iccs,
          expectedMatchVersion: version.value,
        }),
      })
      // 用服务端快照同步状态
      applySnapshot(snapshot)
    } catch {
      // 请求失败，恢复状态
      status.value = prevStatus
      thinking.value = false
      const ui = useUiStore()
      ui.showToast('着法提交失败，请重试')
    }
  }

  // ── 悔棋 ──
  async function undo() {
    if (!matchId.value || !allowUndo.value) return false
    try {
      const snapshot = await apiRequest<MatchSnapshot>(`/matches/${matchId.value}/undo`, {
        method: 'POST',
        body: JSON.stringify({ expectedMatchVersion: version.value }),
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
      const snapshot = await apiRequest<MatchSnapshot>(`/matches/${matchId.value}/resign`, {
        method: 'POST',
        body: JSON.stringify({ expectedMatchVersion: version.value }),
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
      const result = await apiRequest<{ accepted: boolean; match: MatchSnapshot }>(
        `/matches/${matchId.value}/draw-offers`,
        {
          method: 'POST',
          body: JSON.stringify({ expectedMatchVersion: version.value }),
        },
      )
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
    matchId.value = null
    version.value = 0
    status.value = 'active_player_turn'
    playerColor.value = 'red'
    sideToMove.value = 'red'
    difficulty.value = 0
    engine.value = ''
    allowUndo.value = true
    fen.value = initialFEN
    moves.value = []
    outcome.value = 'ongoing'
    drawOffered.value = false
    thinking.value = false
    clearSelection()
  }

  // ── 清理 ──
  function dispose() {
    disconnectStream()
    reset()
  }

  return {
    // 状态
    matchId, version, status, playerColor, sideToMove, difficulty,
    engine, allowUndo, fen, moves, outcome, drawOffered,
    // UI
    flipped, soundEnabled, selectedPos, hints, thinking,
    // 派生
    pieces, myTurn, isFinished, statusLabel, lastMove,
    // 方法
    createMatch, loadMatch, selectPieceAt, clearSelection, submitMove,
    undo, resign, offerDraw, reset, dispose,
  }
})
