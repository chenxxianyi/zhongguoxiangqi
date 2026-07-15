<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import { useAnalysisStore } from '@/stores/analysis'
import { useMatchStore } from '@/stores/match'
import { useUiStore } from '@/stores/ui'
import { iccsToDisplay } from '@/utils/moveNotation'
import type { MoveAnalysis } from '@/api/contracts'

const route = useRoute()
const router = useRouter()
const analysis = useAnalysisStore()
const match = useMatchStore()
const ui = useUiStore()

const selectedMovePly = ref<number | null>(null)

onMounted(async () => {
  const matchId = route.params.matchId as string
  if (!matchId) {
    ui.showToast('未指定对局 ID')
    await router.push('/history')
    return
  }

  // 加载对局信息
  try {
    await match.loadMatch(matchId)
  } catch {
    // 对局不存在也可以分析
  }

  // 加载分析结果
  await analysis.loadOrCreateAnalysis(matchId)
})

onUnmounted(() => {
  match.dispose()
  analysis.reset()
})

// ── 分类中文名 ──
const classLabels: Record<string, string> = {
  best: '最佳着',
  excellent: '优秀',
  inaccuracy: '不精确',
  mistake: '失误',
  blunder: '大漏',
  outside_top_candidates: '非候选',
}

const classColors: Record<string, string> = {
  best: 'best',
  excellent: 'good',
  inaccuracy: 'inaccuracy',
  mistake: 'mistake',
  blunder: 'blunder',
  outside_top_candidates: 'neutral',
}

function formatScore(loss?: number): string {
  if (loss === undefined || loss === null) return '-'
  if (loss === 0) return '0.00'
  return `-${(loss / 100).toFixed(2)}`
}
</script>

<template>
  <section class="page active">
    <!-- 加载状态 -->
    <div v-if="analysis.loading" class="loading-state">
      <p>正在分析对局，请稍候…</p>
      <div class="import-progress" v-if="analysis.currentJob">
        <span :style="{ width: `${analysis.currentJob.progress}%` }" />
      </div>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="analysis.error" class="error-state">
      <p>{{ analysis.error }}</p>
      <button class="secondary-button" @click="router.push('/history')">返回历史对局</button>
    </div>

    <!-- 分析结果 -->
    <template v-else-if="analysis.currentResult">
      <!-- 摘要 -->
      <div class="analysis-summary">
        <div class="analysis-result">
          <span class="result-badge win large">{{ match.outcome === 'red_win' ? '红胜' : match.outcome === 'black_win' ? '黑胜' : '和棋' }}</span>
          <div>
            <span class="section-kicker">对局结果</span>
            <h2>{{ match.engine }} 分析报告</h2>
            <p>共分析 {{ analysis.currentResult.analyzedMoves }} 步着法</p>
          </div>
        </div>
        <div class="analysis-score">
          <div>
            <span>综合表现</span>
            <strong>{{ Math.round(analysis.currentResult.bestMoveRate * 100) }}</strong>
            <small>/ 100</small>
          </div>
          <dl>
            <div><dt>最佳着率</dt><dd>{{ (analysis.currentResult.bestMoveRate * 100).toFixed(0) }}%</dd></div>
            <div><dt>分析引擎</dt><dd>{{ analysis.currentResult.engine }}</dd></div>
          </dl>
        </div>
      </div>

      <!-- 着法列表 -->
      <div class="analysis-layout">
        <div class="analysis-main">
          <article class="surface turning-panel">
            <div class="panel-header">
              <div>
                <span class="section-kicker">逐着分析</span>
                <h3>每步评分</h3>
              </div>
            </div>
            <div v-if="analysis.currentResult.moves.length === 0" class="empty-state">
              暂无分析数据
            </div>
            <button
              v-for="(move, index) in analysis.currentResult.moves"
              :key="move.ply"
              class="turning-row"
              :class="{ active: selectedMovePly === move.ply }"
              @click="selectedMovePly = selectedMovePly === move.ply ? null : move.ply"
            >
              <span class="turn-number">{{ move.ply }}</span>
              <span>
                <strong>{{ iccsToDisplay(move.actualMove, match.fen, move.side as 'red' | 'black') }}</strong>
                <small>最佳：{{ iccsToDisplay(move.bestMove, match.fen, move.side as 'red' | 'black') }}</small>
              </span>
              <span class="turn-quality" :class="classColors[move.classification]">
                {{ classLabels[move.classification] || move.classification }}
              </span>
              <b>{{ formatScore(move.scoreLossCp) }}</b>
            </button>
          </article>
        </div>

        <aside class="analysis-side">
          <article class="surface weakness-panel">
            <span class="section-kicker">分析概要</span>
            <h3>着法统计</h3>
            <div class="pattern-score">
              <span>最佳着</span>
              <strong>{{ analysis.currentResult.moves.filter(m => m.classification === 'best').length }}</strong>
            </div>
            <div class="pattern-score">
              <span>失误</span>
              <strong>{{ analysis.currentResult.moves.filter(m => m.classification === 'mistake' || m.classification === 'blunder').length }}</strong>
            </div>
            <div class="pattern-score">
              <span>分析深度</span>
              <strong>{{ analysis.currentResult.moves[0]?.depth ?? 0 }} 层</strong>
            </div>
          </article>
        </aside>
      </div>
    </template>

    <!-- 无数据 -->
    <div v-else class="empty-state">
      <p>无法加载分析数据</p>
      <button class="secondary-button" @click="router.push('/history')">返回历史对局</button>
    </div>
  </section>
</template>
