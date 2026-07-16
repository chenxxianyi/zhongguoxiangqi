<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import AppModal from '@/components/common/AppModal.vue'
import XiangqiBoard from '@/components/board/XiangqiBoard.vue'
import { useMatchStore } from '@/stores/match'
import { useUiStore } from '@/stores/ui'
import { formatMoveList } from '@/utils/moveNotation'

const route = useRoute()
const router = useRouter()
const match = useMatchStore()
const ui = useUiStore()

const tab = ref<'moves' | 'info'>('moves')
const resignOpen = ref(false)

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
    ui.showToast('无法加载对局，请确认对局 ID 是否正确')
    await router.push('/')
  }
})

onUnmounted(() => {
  match.dispose()
})

// ── 着法列表 ──
const moveRows = ref<Array<[string, string, string]>>([])

watch(
  () => match.moves,
  (moves) => {
    moveRows.value = formatMoveList(moves, match.fen)
  },
  { immediate: true },
)

// ── 对局结束处理 ──
watch(
  () => match.isFinished,
  (finished) => {
    if (finished) {
      const labels: Record<string, string> = {
        red_win: '红方胜',
        black_win: '黑方胜',
        draw: '和棋',
      }
      ui.showToast(`对局结束：${labels[match.outcome] ?? match.outcome}`)
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

// ── 对手信息 ──
const opponentLabel = computed(() => {
  if (match.engine) {
    return `${match.engine} · ${match.difficulty} 级`
  }
  return `棋境 AI · ${match.difficulty} 级`
})

const aiModeLabel = computed(() =>
  ({ standard: '标准引擎', library: '棋谱库优先', style: '棋风模仿' })[match.aiMode],
)
</script>

<template>
  <section class="page active match-page">
    <div class="match-shell">
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
          <time v-if="match.moves.length > 0">{{ match.moves.length }} 步</time>
        </div>

        <XiangqiBoard @undo="handleUndo" @resign="resignOpen = true">
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
            :aria-selected="tab === 'moves'"
            @click="tab = 'moves'"
          >着法</button>
          <button
            :class="{ active: tab === 'info' }"
            role="tab"
            :aria-selected="tab === 'info'"
            @click="tab = 'info'"
          >局面</button>
        </div>

        <!-- 着法列表 -->
        <div v-if="tab === 'moves'" class="match-tab active">
          <ol class="move-list">
            <li
              v-for="(move, index) in moveRows"
              :key="move[0]"
              :class="{ current: index === moveRows.length - 1 }"
            >
              <span>{{ move[0] }}</span>
              <button>{{ move[1] }}</button>
              <button>{{ move[2] }}</button>
            </li>
          </ol>
          <div v-if="moveRows.length === 0" class="move-list-empty">
            <AppIcon name="board" />
            <p>对局尚未开始，等待着你走出第一步。</p>
          </div>
        </div>

        <!-- 局面信息 -->
        <div v-else class="match-tab active">
          <dl class="position-details">
            <div><dt>对局编号</dt><dd>{{ match.matchId?.slice(0, 12) ?? '-' }}…</dd></div>
            <div><dt>当前 FEN</dt><dd>{{ match.fen.split(' ')[0] ?? match.fen }}</dd></div>
            <div><dt>对局版本</dt><dd>v{{ match.version }}</dd></div>
            <div><dt>当前行棋方</dt><dd>{{ match.sideToMove === 'red' ? '红方' : '黑方' }}</dd></div>
            <div><dt>执色</dt><dd>{{ match.playerColor === 'red' ? '红方（先手）' : '黑方（后手）' }}</dd></div>
            <div><dt>AI 模式</dt><dd>{{ aiModeLabel }}</dd></div>
            <div><dt>引擎</dt><dd>{{ match.engine || '内置引擎' }}</dd></div>
            <div><dt>AI 思考中</dt><dd>{{ match.thinking ? '是' : '否' }}</dd></div>
            <div v-if="match.isFinished">
              <dt>结果</dt>
              <dd>
                <span class="result-badge" :class="{
                  win: match.outcome === 'red_win' && match.playerColor === 'red' || match.outcome === 'black_win' && match.playerColor === 'black',
                  loss: match.outcome === 'red_win' && match.playerColor === 'black' || match.outcome === 'black_win' && match.playerColor === 'red',
                  draw: match.outcome === 'draw',
                }">
                  {{ match.outcome === 'red_win' ? '红胜' : match.outcome === 'black_win' ? '黑胜' : '和棋' }}
                </span>
              </dd>
            </div>
          </dl>
        </div>

        <!-- 操作按钮 -->
        <div class="match-actions">
          <button
            v-if="match.isFinished"
            class="secondary-button full"
            @click="router.push(`/analysis/${match.matchId}`)"
          >
            <AppIcon name="chart" />查看复盘
          </button>
          <button
            v-if="!match.isFinished && match.myTurn"
            class="secondary-button full"
            @click="handleDraw"
          >
            <AppIcon name="eye" />请求和棋
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
