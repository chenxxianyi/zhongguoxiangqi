<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import AppModal from '@/components/common/AppModal.vue'
import XiangqiBoard from '@/components/board/XiangqiBoard.vue'
import { useMatchStore } from '@/stores/match'
import { useUiStore } from '@/stores/ui'
import { formatMoveList } from '@/utils/moveNotation'
import { getCapturedPieceLabel } from '@/utils/capturedPiece'
import {
  getPlayerResult,
  getPlayerResultClass,
  getPlayerResultLabel,
  getTerminationLabel,
} from '@/utils/matchResult'

const route = useRoute()
const router = useRouter()
const match = useMatchStore()
const ui = useUiStore()

const tab = ref<'moves' | 'info'>('moves')
const resignOpen = ref(false)
const loading = ref(true)
const loadError = ref<string | null>(null)
const moveList = ref<HTMLOListElement | null>(null)

// ── 加载对局 ──
onMounted(async () => {
  const id = route.params.id as string
  if (!id) {
    await router.push('/new-game')
    return
  }
  try {
    await match.loadMatch(id)
  } catch {
    loadError.value = '无法从后端加载该对局'
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  ui.matchFocusMode = false
  match.dispose()
})

// ── 着法列表 ──
const moveRows = ref<Array<[string, string, string]>>([])

watch(
  () => match.moves,
  (moves) => {
    moveRows.value = formatMoveList(moves, match.fen)
    void nextTick(() => {
      moveList.value?.lastElementChild?.scrollIntoView({ block: 'nearest' })
    })
  },
  { immediate: true },
)

// ── 对局结束处理 ──
watch(
  () => match.isFinished,
  (finished) => {
    if (finished) {
      const result = getPlayerResult(match.outcome, match.playerColor)
      const label = getPlayerResultLabel(result)
      const reason = getTerminationLabel(match.termination)
      ui.showToast(`对局结束：${result === 'draw' ? '和棋' : `你${label}`} · ${reason}`)
    }
  },
)

// ── 操作 ──
async function handleUndo() {
  await match.undo()
}

async function confirmResign() {
  resignOpen.value = false
  const ok = await match.resign()
  if (ok) {
    ui.showToast('本局已结束')
  }
}

async function handleDraw() {
  await match.offerDraw()
}

function toggleFocusMode() {
  ui.matchFocusMode = !ui.matchFocusMode
  ui.showToast(ui.matchFocusMode ? '已进入专注模式' : '已退出专注模式')
}

function handlePanelTabKeydown(event: KeyboardEvent) {
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
  event.preventDefault()
  tab.value = event.key === 'ArrowLeft' || event.key === 'Home' ? 'moves' : 'info'
  void nextTick(() => document.getElementById(`match-tab-${tab.value}`)?.focus())
}

function moveStateLabel(moveIndex: number) {
  const move = match.moves[moveIndex]
  if (!move?.givesCheck) return ''
  const isFinalMove = moveIndex === match.moves.length - 1
  return isFinalMove && match.termination === 'checkmate' ? '杀' : '将'
}

function moveCaptureLabel(moveIndex: number) {
  const label = getCapturedPieceLabel(match.moves[moveIndex]?.captured)
  return label ? `吃${label}` : ''
}

// ── 对手信息 ──
const opponentLabel = computed(() => {
  if (match.engine) {
    return `${match.engine} · ${match.difficulty} 级`
  }
  return '引擎信息未提供'
})

const aiModeLabel = computed(() =>
  ({ standard: '标准引擎', library: '棋谱库优先', style: '棋风模仿' })[match.aiMode],
)
const playerResult = computed(() => getPlayerResult(match.outcome, match.playerColor))
</script>

<template>
  <section class="page active match-page">
    <div v-if="loading" class="loading-state">正在读取后端对局…</div>
    <div v-else-if="loadError" class="error-state">
      <p>{{ loadError }}</p>
      <button class="secondary-button" @click="router.push('/history')">返回历史对局</button>
    </div>
    <div v-else class="match-shell">
      <div class="match-board-column">
        <!-- 对手栏 -->
        <div class="player-bar opponent">
          <div class="player-identity">
            <span class="avatar ai">AI</span>
            <div>
              <strong>棋境 AI</strong>
              <small>{{ opponentLabel }}</small>
            </div>
          </div>
          <div v-if="match.thinking" class="thinking-label">
            <span class="thinking-dot" />正在思考
          </div>
          <div class="match-player-actions">
            <time v-if="match.moves.length > 0">{{ match.moves.length }} 手</time>
            <button
              class="icon-button match-focus-button"
              :class="{ active: ui.matchFocusMode }"
              :aria-pressed="ui.matchFocusMode"
              :aria-label="ui.matchFocusMode ? '退出专注模式' : '进入专注模式'"
              @click="toggleFocusMode"
            ><AppIcon name="focus" /></button>
          </div>
        </div>

        <XiangqiBoard
          @undo="handleUndo"
          @resign="resignOpen = true"
          @restart="router.push('/new-game')"
          @review="router.push(`/analysis/${match.matchId}`)"
        >
          <!-- 己方栏 -->
          <div class="player-bar self">
            <div class="player-identity">
              <span class="avatar">你</span>
              <div>
                <strong>你</strong>
                <small>{{ match.playerColor === 'red' ? '红方' : '黑方' }} · {{ match.statusLabel }}</small>
              </div>
            </div>
            <div v-if="match.myTurn" class="turn-label">你的回合</div>
          </div>
        </XiangqiBoard>
      </div>

      <aside class="match-panel surface">
        <div class="match-panel-tabs" role="tablist">
          <button
            :class="{ active: tab === 'moves' }"
            role="tab"
            id="match-tab-moves"
            aria-controls="match-panel-moves"
            :aria-selected="tab === 'moves'"
            :tabindex="tab === 'moves' ? 0 : -1"
            @click="tab = 'moves'"
            @keydown="handlePanelTabKeydown"
          >着法</button>
          <button
            :class="{ active: tab === 'info' }"
            role="tab"
            id="match-tab-info"
            aria-controls="match-panel-info"
            :aria-selected="tab === 'info'"
            :tabindex="tab === 'info' ? 0 : -1"
            @click="tab = 'info'"
            @keydown="handlePanelTabKeydown"
          >局面</button>
        </div>

        <!-- 着法列表 -->
        <div
          v-if="tab === 'moves'"
          id="match-panel-moves"
          class="match-tab active"
          role="tabpanel"
          aria-labelledby="match-tab-moves"
        >
          <ol ref="moveList" class="move-list">
            <li
              v-for="(move, index) in moveRows"
              :key="move[0]"
              :class="{ current: index === moveRows.length - 1 }"
            >
              <span>{{ move[0] }}</span>
              <button>
                {{ move[1] }}
                <span v-if="moveCaptureLabel(index * 2)" class="move-capture-mark">{{ moveCaptureLabel(index * 2) }}</span>
                <span v-if="moveStateLabel(index * 2)" class="move-state-mark">{{ moveStateLabel(index * 2) }}</span>
              </button>
              <button>
                {{ move[2] }}
                <span v-if="moveCaptureLabel(index * 2 + 1)" class="move-capture-mark">{{ moveCaptureLabel(index * 2 + 1) }}</span>
                <span v-if="moveStateLabel(index * 2 + 1)" class="move-state-mark">{{ moveStateLabel(index * 2 + 1) }}</span>
              </button>
            </li>
          </ol>
          <div v-if="moveRows.length === 0" class="move-list-empty">
            <AppIcon name="board" />
            <p>对局尚未开始，等待着你走出第一步。</p>
          </div>
        </div>

        <!-- 局面信息 -->
        <div
          v-else
          id="match-panel-info"
          class="match-tab active"
          role="tabpanel"
          aria-labelledby="match-tab-info"
        >
          <dl class="position-details">
            <div><dt>对局编号</dt><dd>{{ match.matchId?.slice(0, 12) ?? '-' }}…</dd></div>
            <div><dt>当前 FEN</dt><dd>{{ match.fen.split(' ')[0] ?? match.fen }}</dd></div>
            <div><dt>对局版本</dt><dd>v{{ match.version }}</dd></div>
            <div><dt>当前行棋方</dt><dd>{{ match.sideToMove === 'red' ? '红方' : '黑方' }}</dd></div>
            <div><dt>局面状态</dt><dd :class="{ 'danger-text': match.inCheck }">{{ match.inCheck ? '将军' : '正常' }}</dd></div>
            <div><dt>执色</dt><dd>{{ match.playerColor === 'red' ? '红方（先手）' : '黑方（后手）' }}</dd></div>
            <div><dt>AI 模式</dt><dd>{{ aiModeLabel }}</dd></div>
            <div><dt>引擎</dt><dd>{{ match.engine || '未提供' }}</dd></div>
            <div><dt>AI 思考中</dt><dd>{{ match.thinking ? '是' : '否' }}</dd></div>
            <div v-if="match.isFinished">
              <dt>结果</dt>
              <dd>
                <span class="result-badge" :class="getPlayerResultClass(playerResult)">
                  {{ getPlayerResultLabel(playerResult) }}
                </span>
              </dd>
            </div>
            <div v-if="match.isFinished"><dt>终局原因</dt><dd>{{ getTerminationLabel(match.termination) }}</dd></div>
          </dl>
        </div>

        <!-- 操作按钮 -->
        <div class="match-actions">
          <div v-if="match.isFinished" class="match-finish-actions">
            <button class="primary-button full" @click="router.push('/new-game')">
              <AppIcon name="refresh" />再来一局
            </button>
            <button class="secondary-button full" @click="router.push(`/analysis/${match.matchId}`)">
              <AppIcon name="chart" />查看复盘
            </button>
          </div>
          <button
            v-if="!match.isFinished && match.myTurn"
            class="secondary-button full"
            @click="handleDraw"
          >
            <AppIcon name="handshake" />请求和棋
          </button>
        </div>
      </aside>
    </div>

    <AppModal
      :open="resignOpen"
      title="确认认输？"
      description="本局将立即结束。"
      danger
      @close="resignOpen = false"
      @confirm="confirmResign"
    />
  </section>
</template>
