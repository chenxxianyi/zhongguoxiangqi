<script setup lang="ts">
import { useRouter } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import { historyMatches } from '@/data/demo'
const router = useRouter()
</script>

<template>
  <section class="page active"><div class="section-intro split"><div><span class="section-kicker">对局档案</span><h2>回看你的棋路</h2><p>按执色、难度、AI 模式和复盘状态筛选。</p></div><button class="secondary-button"><AppIcon name="download" />导出我的数据</button></div><div class="history-stats"><article><span>总对局</span><strong>126</strong><small>本月 24 盘</small></article><article><span>胜 / 和 / 负</span><strong><i class="red-text">63</i> / 18 / 45</strong><small>总胜率 50%</small></article><article><span>最常难度</span><strong>进阶 6</strong><small>共 48 盘</small></article><article><span>平均用时</span><strong>18:42</strong><small>平均 46 回合</small></article></div>
  <article class="surface history-list-panel"><div class="records-toolbar"><div class="search-field"><AppIcon name="eye" /><input type="search" placeholder="搜索开局或对局编号" aria-label="搜索历史对局"></div><button class="filter-button"><AppIcon name="filter" />全部结果</button><button class="filter-button"><AppIcon name="filter" />全部难度</button></div><div class="history-list"><button v-for="match in historyMatches" :key="`${match.opponent}-${match.date}`" class="history-row" @click="router.push('/analysis')"><span class="result-badge" :class="match.resultClass">{{ match.result }}</span><span><strong>{{ match.opponent }}</strong><small>{{ match.opening }}</small></span><span><strong>{{ match.mode }}</strong><small>AI 模式</small></span><span><strong>{{ match.length }}</strong><small>对局长度</small></span><span><strong>{{ match.bookHit }}</strong><small>学习命中</small></span><span class="tag success">已复盘</span><AppIcon class="row-chevron" name="chevron" /></button></div></article></section>
</template>
