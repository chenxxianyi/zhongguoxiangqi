<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import { apiRequest } from '@/api/client'
import type { MatchSnapshot } from '@/api/contracts'

const router = useRouter()
const matches = ref<MatchSnapshot[]>([])
const loaded = ref(false)

// ── 统计 ──
const totalGames = ref(0)
const wins = ref(0)
const draws = ref(0)
const losses = ref(0)
const winRateDisplay = ref('0%')

onMounted(async () => {
  try {
    const result = await apiRequest<{ items: MatchSnapshot[] }>('/matches')
    matches.value = result.items
    totalGames.value = result.items.length

    const finished = result.items.filter((m) => m.status === 'finished')
    wins.value = finished.filter((m) => {
      return (m.playerColor === 'red' && m.outcome === 'red_win') ||
             (m.playerColor === 'black' && m.outcome === 'black_win')
    }).length
    draws.value = finished.filter((m) => m.outcome === 'draw').length
    losses.value = finished.length - wins.value - draws.value

    if (finished.length > 0) {
      winRateDisplay.value = Math.round((wins.value / finished.length) * 100) + '%'
    }

    loaded.value = true
  } catch {
    console.warn('无法获取对局列表')
    loaded.value = true
  }
})

function outcomeClass(outcome: string, playerColor: string): string {
  if (outcome === 'draw') return 'draw'
  const win = (outcome === 'red_win' && playerColor === 'red') || (outcome === 'black_win' && playerColor === 'black')
  return win ? 'win' : 'loss'
}

function outcomeLabel(outcome: string): string {
  return outcome === 'red_win' ? '胜' : outcome === 'black_win' ? '负' : '和'
}

function formatDate(dateStr: string): string {
  try {
    const d = new Date(dateStr)
    const now = new Date()
    const diffDays = Math.floor((now.getTime() - d.getTime()) / (1000 * 60 * 60 * 24))
    if (diffDays === 0) return '今日'
    if (diffDays === 1) return '昨日'
    if (diffDays < 30) return `${diffDays} 天前`
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  } catch {
    return dateStr.slice(0, 10)
  }
}

function statusTag(status: string): string {
  if (status === 'finished') return '已结束'
  if (status === 'active_player_turn' || status === 'active_ai_thinking') return '进行中'
  return status
}
</script>

<template>
  <section class="page active">
    <div class="section-intro split">
      <div>
        <span class="section-kicker">对局档案</span>
        <h2>回看你的棋路</h2>
        <p>按执色、难度和结果查看历史对局。</p>
      </div>
    </div>

    <div class="history-stats" v-if="loaded">
      <article>
        <span>总对局</span>
        <strong>{{ totalGames }}</strong>
        <small>全部记录</small>
      </article>
      <article>
        <span>胜 / 和 / 负</span>
        <strong>
          <i class="red-text">{{ wins }}</i>
          / {{ draws }} / {{ losses }}
        </strong>
        <small>胜率 {{ winRateDisplay }}</small>
      </article>
    </div>

    <article class="surface history-list-panel">
      <div v-if="!loaded" class="loading-state">加载中…</div>
      <div v-else-if="matches.length === 0" class="empty-state">
        暂无对局记录。前往<a href="/new-game">新建对局</a>开始下棋。
      </div>
      <div v-else class="history-list">
        <button
          v-for="matchItem in matches"
          :key="matchItem.id"
          class="history-row"
          @click="router.push(`/analysis/${matchItem.id}`)"
        >
          <span class="result-badge" :class="outcomeClass(matchItem.outcome, matchItem.playerColor)">
            {{ outcomeLabel(matchItem.outcome) }}
          </span>
          <span>
            <strong>AI · {{ matchItem.difficulty }} 级</strong>
            <small>{{ matchItem.playerColor === 'red' ? '执红' : '执黑' }}</small>
          </span>
          <span>
            <strong>{{ matchItem.moves?.length ?? 0 }} 步</strong>
            <small>对局长度</small>
          </span>
          <span class="tag" :class="matchItem.status === 'finished' ? 'neutral' : 'success'">
            {{ statusTag(matchItem.status) }}
          </span>
          <span class="history-date">{{ formatDate(matchItem.createdAt) }}</span>
          <AppIcon class="row-chevron" name="chevron" />
        </button>
      </div>
    </article>
  </section>
</template>
