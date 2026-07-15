<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import MiniBoard from '@/components/board/MiniBoard.vue'
import { apiRequest } from '@/api/client'
import type { MatchSnapshot } from '@/api/contracts'

const matches = ref<MatchSnapshot[]>([])
const recentMatches = ref<MatchSnapshot[]>([])
const loaded = ref(false)

// ── 统计 ──
const totalMatches = ref(0)
const winRate = ref(0)
const recentCount = ref(0)
const previousMonthCount = ref(0)

onMounted(async () => {
  try {
    const result = await apiRequest<{ items: MatchSnapshot[] }>('/matches')
    matches.value = result.items
    recentMatches.value = result.items.slice(0, 5)

    // 计算统计
    totalMatches.value = result.items.length

    const now = Date.now()
    const thirtyDaysAgo = now - 30 * 24 * 60 * 60 * 1000
    const sixtyDaysAgo = now - 60 * 24 * 60 * 60 * 1000

    const recent = result.items.filter((m) => new Date(m.createdAt).getTime() > thirtyDaysAgo)
    const previous = result.items.filter(
      (m) => new Date(m.createdAt).getTime() > sixtyDaysAgo && new Date(m.createdAt).getTime() <= thirtyDaysAgo,
    )

    recentCount.value = recent.length
    previousMonthCount.value = previous.length

    // 胜率
    const finished = recent.filter((m) => m.status === 'finished')
    if (finished.length > 0) {
      const wins = finished.filter((m) => {
        const isPlayerRed = m.playerColor === 'red'
        return (isPlayerRed && m.outcome === 'red_win') || (!isPlayerRed && m.outcome === 'black_win')
      })
      winRate.value = Math.round((wins.length / finished.length) * 100)
    }

    loaded.value = true
  } catch {
    // 后端不可用时静默降级
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

function formatMatchDate(dateStr: string): string {
  try {
    const d = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - d.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
    if (diffDays === 0) return '今日'
    if (diffDays === 1) return '昨日'
    if (diffDays < 7) return `${diffDays} 天前`
    return `${d.getMonth() + 1} 月 ${d.getDate()} 日`
  } catch {
    return dateStr.slice(0, 10)
  }
}
</script>

<template>
  <section class="page active">
    <div class="hero-panel">
      <div class="hero-copy">
        <span class="section-kicker">继续你的棋局</span>
        <h2>在每一次落子中，<br><em>看见自己的进步。</em></h2>
        <p>与可调节难度的 AI 对弈，导入经典棋谱，建立属于你的棋风学习库。</p>
        <div class="hero-actions">
          <RouterLink class="primary-button" to="/new-game"><AppIcon name="play" />新建对局</RouterLink>
          <RouterLink v-if="matches.length > 0" class="text-button" :to="`/match/${matches[0]!.id}`">继续上局 <AppIcon name="arrow" /></RouterLink>
        </div>
      </div>
      <div class="hero-board-wrap" aria-hidden="true"><MiniBoard /><div class="hero-stamp">棋境<br>待续</div></div>
    </div>

    <div class="metric-strip" aria-label="近期统计" v-if="loaded">
      <article>
        <span class="metric-icon cinnabar"><AppIcon name="board" /></span>
        <div>
          <span>近 30 日对局</span>
          <strong>{{ recentCount }}</strong>
          <small>{{ previousMonthCount > 0 ? `较上月 ${recentCount - previousMonthCount >= 0 ? '+' : ''}${recentCount - previousMonthCount}` : '暂无对比' }}</small>
        </div>
      </article>
      <article>
        <span class="metric-icon jade"><AppIcon name="chart" /></span>
        <div>
          <span>当前胜率</span>
          <strong>{{ winRate }}%</strong>
          <small>{{ totalMatches > 0 ? `共 ${totalMatches} 盘` : '暂无对局' }}</small>
        </div>
      </article>
      <article>
        <span class="metric-icon ochre"><AppIcon name="spark" /></span>
        <div>
          <span>总对局</span>
          <strong>{{ totalMatches }}</strong>
          <small>持续增长中</small>
        </div>
      </article>
    </div>

    <div class="dashboard-grid">
      <article class="surface recent-match-panel">
        <div class="panel-header">
          <div><span class="section-kicker">最近对局</span><h3>复盘你的每一步</h3></div>
          <RouterLink v-if="matches.length > 0" class="text-button small" to="/history">查看全部 <AppIcon name="arrow" /></RouterLink>
        </div>
        <div v-if="recentMatches.length === 0" class="empty-state">
          还没有对局记录，开始你的第一盘棋吧。
        </div>
        <div v-else class="match-list">
          <RouterLink
            v-for="matchItem in recentMatches"
            :key="matchItem.id"
            class="match-row"
            :to="`/match/${matchItem.id}`"
          >
            <span class="result-badge" :class="outcomeClass(matchItem.outcome, matchItem.playerColor)">
              {{ outcomeLabel(matchItem.outcome) }}
            </span>
            <span class="match-opponent">
              <strong>AI · {{ matchItem.difficulty }} 级</strong>
              <small>{{ matchItem.playerColor === 'red' ? '执红' : '执黑' }}</small>
            </span>
            <span class="match-score">
              <strong>{{ matchItem.moves?.length ?? 0 }} 步</strong>
              <small>{{ formatMatchDate(matchItem.createdAt) }}</small>
            </span>
            <span v-if="matchItem.status === 'active_player_turn' || matchItem.status === 'active_ai_thinking'" class="tag neutral">进行中</span>
            <AppIcon class="row-chevron" name="chevron" />
          </RouterLink>
        </div>
      </article>

      <article class="surface learning-summary">
        <div class="panel-header">
          <div><span class="section-kicker">学习库</span><h3>导入棋谱，提升棋力</h3></div>
          <span class="tag neutral">等待数据</span>
        </div>
        <div class="library-meta">
          <div><span>导入棋谱</span><strong>0 盘</strong></div>
          <div><span>学习版本</span><strong>0</strong></div>
        </div>
        <RouterLink class="secondary-button full" to="/records">导入棋谱</RouterLink>
      </article>
    </div>
  </section>
</template>
