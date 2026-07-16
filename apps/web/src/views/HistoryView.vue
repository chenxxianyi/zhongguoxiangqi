<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import { listMatches } from '@/api/matches'
import { formatRelativeDate } from '@/utils/date'
import {
  getMatchDestination,
  getMatchResult,
  getMatchResultStats,
  getPlayerResultClass,
  getPlayerResultLabel,
  isActiveMatch,
} from '@/utils/matchResult'
import type { MatchSnapshot, MatchStatus } from '@/api/contracts'

const router = useRouter()
const matches = ref<MatchSnapshot[]>([])
const loaded = ref(false)
const error = ref<string | null>(null)

// ── 统计 ──
const stats = computed(() => getMatchResultStats(matches.value))
const totalGames = computed(() => stats.value.total)
const wins = computed(() => stats.value.wins)
const draws = computed(() => stats.value.draws)
const losses = computed(() => stats.value.losses)
const winRateDisplay = computed(() => `${stats.value.winRate}%`)

onMounted(async () => {
  try {
    matches.value = await listMatches()
  } catch {
    error.value = '无法从后端获取历史对局'
  } finally {
    loaded.value = true
  }
})

function statusTag(status: MatchStatus): string {
  if (status === 'finished') return '已结束'
  if (isActiveMatch(status)) return '进行中'
  if (status === 'aborted') return '已中止'
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

    <div v-if="loaded && !error" class="history-stats">
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
      <div v-else-if="error" class="error-state">
        <p>{{ error }}</p>
      </div>
      <div v-else-if="matches.length === 0" class="empty-state">
        暂无对局记录。前往<a href="/new-game">新建对局</a>开始下棋。
      </div>
      <div v-else class="history-list">
        <button
          v-for="matchItem in matches"
          :key="matchItem.id"
          class="history-row"
          @click="router.push(getMatchDestination(matchItem))"
        >
            <span
              class="result-badge"
              :class="getPlayerResultClass(getMatchResult(matchItem))"
            >
            {{ getPlayerResultLabel(getMatchResult(matchItem)) }}
          </span>
          <span>
            <strong>AI · {{ matchItem.difficulty }} 级</strong>
            <small>{{ matchItem.playerColor === 'red' ? '执红' : '执黑' }}</small>
          </span>
          <span>
            <strong>{{ matchItem.moves?.length ?? 0 }} 步</strong>
            <small>对局长度</small>
          </span>
          <span class="tag" :class="isActiveMatch(matchItem.status) ? 'success' : 'neutral'">
            {{ statusTag(matchItem.status) }}
          </span>
          <span class="history-date">{{ formatRelativeDate(matchItem.createdAt) }}</span>
          <AppIcon class="row-chevron" name="chevron" />
        </button>
      </div>
    </article>
  </section>
</template>
