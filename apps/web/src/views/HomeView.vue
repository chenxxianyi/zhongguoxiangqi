<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import MiniBoard from '@/components/board/MiniBoard.vue'
import { listLearningVersions } from '@/api/learning'
import { listMatches } from '@/api/matches'
import { listRecords } from '@/api/records'
import { getPiecesFromFEN } from '@/utils/board'
import { formatRelativeDate } from '@/utils/date'
import {
  getMatchDestination,
  getMatchResult,
  getMatchResultStats,
  getPlayerResultClass,
  getPlayerResultLabel,
  isActiveMatch,
} from '@/utils/matchResult'
import type { LearningVersion, MatchSnapshot } from '@/api/contracts'

const matches = ref<MatchSnapshot[]>([])
const importedRecordCount = ref(0)
const learningVersions = ref<LearningVersion[]>([])
const loaded = ref(false)
const matchesUnavailable = ref(false)
const recordsUnavailable = ref(false)
const versionsUnavailable = ref(false)

const recentMatches = computed(() => matches.value.slice(0, 5))
const latestMatchPieces = computed(() => {
  const latestMatch = matches.value[0]
  return latestMatch ? getPiecesFromFEN(latestMatch.fen) : []
})
const activeLearningVersion = computed(() =>
  learningVersions.value.find((version) => version.status === 'active') ?? null,
)
const learningStatusLabel = computed(() => {
  if (versionsUnavailable.value) return '服务不可用'
  if (activeLearningVersion.value) return '已启用'
  if (learningVersions.value.length > 0) return '待启用'
  return '暂无版本'
})

const totalMatches = computed(() => matches.value.length)
const recentMatchesForStats = computed(() => {
  const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000
  return matches.value.filter((match) => new Date(match.createdAt).getTime() > thirtyDaysAgo)
})
const previousMonthCount = computed(() => {
  const now = Date.now()
  const thirtyDaysAgo = now - 30 * 24 * 60 * 60 * 1000
  const sixtyDaysAgo = now - 60 * 24 * 60 * 60 * 1000
  return matches.value.filter((match) => {
    const createdAt = new Date(match.createdAt).getTime()
    return createdAt > sixtyDaysAgo && createdAt <= thirtyDaysAgo
  }).length
})
const recentCount = computed(() => recentMatchesForStats.value.length)
const winRate = computed(() => getMatchResultStats(recentMatchesForStats.value).winRate)

async function loadDashboard() {
  loaded.value = false
  matchesUnavailable.value = false
  recordsUnavailable.value = false
  versionsUnavailable.value = false

  const [matchResult, recordResult, versionResult] = await Promise.allSettled([
    listMatches(),
    listRecords(),
    listLearningVersions(),
  ])

  if (matchResult.status === 'fulfilled') {
    matches.value = matchResult.value
  } else {
    matchesUnavailable.value = true
  }
  if (recordResult.status === 'fulfilled') {
    importedRecordCount.value = recordResult.value.length
  } else {
    recordsUnavailable.value = true
  }
  if (versionResult.status === 'fulfilled') {
    learningVersions.value = versionResult.value
  } else {
    versionsUnavailable.value = true
  }
  loaded.value = true
}

onMounted(loadDashboard)
</script>

<template>
  <section class="page active">
    <div class="hero-panel">
      <div class="hero-copy">
        <span class="section-kicker">继续你的棋局</span>
        <h2>在每一次落子中，<br><em>看见自己的进步。</em></h2>
        <p>与可调节难度的 AI 对弈，导入经典棋谱，建立属于你的棋风学习库。</p>
        <div class="hero-actions">
          <RouterLink class="primary-button" to="/new-game">
            <AppIcon name="play" />新建对局
          </RouterLink>
          <RouterLink
            v-if="!matchesUnavailable && matches.length > 0"
            class="text-button"
            :to="getMatchDestination(matches[0]!)"
          >
            查看上局 <AppIcon name="arrow" />
          </RouterLink>
        </div>
      </div>
      <div class="hero-board-wrap" aria-hidden="true">
        <MiniBoard :pieces="latestMatchPieces" />
        <div class="hero-stamp">棋境<br>待续</div>
      </div>
    </div>

    <div v-if="loaded" class="metric-strip" aria-label="近期统计">
      <article>
        <span class="metric-icon cinnabar"><AppIcon name="board" /></span>
        <div>
          <span>近 30 日对局</span>
          <strong>{{ matchesUnavailable ? '--' : recentCount }}</strong>
          <small v-if="matchesUnavailable">服务不可用</small>
          <small v-else>
            {{ previousMonthCount > 0 ? `较上月 ${recentCount - previousMonthCount >= 0 ? '+' : ''}${recentCount - previousMonthCount}` : '暂无对比' }}
          </small>
        </div>
      </article>
      <article>
        <span class="metric-icon jade"><AppIcon name="chart" /></span>
        <div>
          <span>当前胜率</span>
          <strong>{{ matchesUnavailable ? '--' : `${winRate}%` }}</strong>
          <small>{{ matchesUnavailable ? '服务不可用' : totalMatches > 0 ? `共 ${totalMatches} 盘` : '暂无对局' }}</small>
        </div>
      </article>
      <article>
        <span class="metric-icon ochre"><AppIcon name="spark" /></span>
        <div>
          <span>总对局</span>
          <strong>{{ matchesUnavailable ? '--' : totalMatches }}</strong>
          <small>{{ matchesUnavailable ? '服务不可用' : '后端累计记录' }}</small>
        </div>
      </article>
    </div>

    <div class="dashboard-grid">
      <article class="surface recent-match-panel">
        <div class="panel-header">
          <div><span class="section-kicker">最近对局</span><h3>复盘你的每一步</h3></div>
          <RouterLink
            v-if="!matchesUnavailable && matches.length > 0"
            class="text-button small"
            to="/history"
          >
            查看全部 <AppIcon name="arrow" />
          </RouterLink>
        </div>
        <div v-if="!loaded" class="loading-state">正在读取后端对局…</div>
        <div v-else-if="matchesUnavailable" class="error-state">
          <p>无法读取后端对局数据。</p>
          <button class="secondary-button small" @click="loadDashboard">重新加载</button>
        </div>
        <div v-else-if="recentMatches.length === 0" class="empty-state">
          还没有对局记录，开始你的第一盘棋吧。
        </div>
        <div v-else class="match-list">
          <RouterLink
            v-for="matchItem in recentMatches"
            :key="matchItem.id"
            class="match-row"
            :to="getMatchDestination(matchItem)"
          >
            <span
              class="result-badge"
              :class="getPlayerResultClass(getMatchResult(matchItem))"
            >
              {{ getPlayerResultLabel(getMatchResult(matchItem)) }}
            </span>
            <span class="match-opponent">
              <strong>AI · {{ matchItem.difficulty }} 级</strong>
              <small>{{ matchItem.playerColor === 'red' ? '执红' : '执黑' }}</small>
            </span>
            <span class="match-score">
              <strong>{{ matchItem.moves?.length ?? 0 }} 步</strong>
              <small>{{ formatRelativeDate(matchItem.createdAt) }}</small>
            </span>
            <span v-if="isActiveMatch(matchItem.status)" class="tag neutral">进行中</span>
            <AppIcon class="row-chevron" name="chevron" />
          </RouterLink>
        </div>
      </article>

      <article class="surface learning-summary">
        <div class="panel-header">
          <div><span class="section-kicker">学习库</span><h3>导入棋谱，提升棋力</h3></div>
          <span class="tag" :class="activeLearningVersion ? 'success' : 'neutral'">
            {{ learningStatusLabel }}
          </span>
        </div>
        <div class="library-meta">
          <div>
            <span>导入棋谱</span>
            <strong>{{ recordsUnavailable ? '--' : `${importedRecordCount} 盘` }}</strong>
            <small v-if="recordsUnavailable">服务不可用</small>
          </div>
          <div>
            <span>学习版本</span>
            <strong>{{ versionsUnavailable ? '--' : learningVersions.length }}</strong>
            <small v-if="versionsUnavailable">服务不可用</small>
          </div>
        </div>
        <RouterLink class="secondary-button full" :to="activeLearningVersion ? '/learning' : '/records'">
          {{ activeLearningVersion ? '查看学习版本' : '导入棋谱' }}
        </RouterLink>
      </article>
    </div>
  </section>
</template>
