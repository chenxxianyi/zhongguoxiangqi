<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import { listMatches } from '@/api/matches'
import { listRecords } from '@/api/records'
import { getEngineHealth } from '@/api/system'
import { useUiStore } from '@/stores/ui'
import { isActiveMatch } from '@/utils/matchResult'
import type { MatchSnapshot } from '@/api/contracts'
import type { EngineHealth } from '@/api/system'

const route = useRoute()
const ui = useUiStore()
const recordCount = ref<number | null>(null)
const activeMatch = ref<MatchSnapshot | null>(null)
const engineHealth = ref<EngineHealth | null>(null)
const engineUnavailable = ref(false)

const navigation = computed(() => [
  { label: '首页', icon: 'grid', to: '/' },
  { label: '新对局', icon: 'play', to: '/new-game' },
  ...(activeMatch.value
    ? [{ label: '当前对局', icon: 'board', to: `/match/${activeMatch.value.id}`, live: true }]
    : []),
  { label: '历史对局', icon: 'clock', to: '/history' },
])

const learningNavigation = [
  { label: '棋谱库', icon: 'record', to: '/records' },
  { label: '学习中心', icon: 'spark', to: '/learning' },
]

const engineStatusLabel = computed(() => {
  if (engineUnavailable.value) return '后端不可用'
  if (!engineHealth.value) return '正在检查'
  return engineHealth.value.status === 'healthy' ? '运行正常' : engineHealth.value.status
})

onMounted(async () => {
  const [recordsResult, matchesResult, engineResult] = await Promise.allSettled([
    listRecords(),
    listMatches(),
    getEngineHealth(),
  ])

  if (recordsResult.status === 'fulfilled') {
    recordCount.value = recordsResult.value.length
  }
  if (matchesResult.status === 'fulfilled') {
    activeMatch.value = matchesResult.value.find((item) => isActiveMatch(item.status)) ?? null
  }
  if (engineResult.status === 'fulfilled') {
    engineHealth.value = engineResult.value
  } else {
    engineUnavailable.value = true
  }
})

function close() {
  ui.sidebarOpen = false
}
</script>

<template>
  <div class="mobile-scrim" :hidden="!ui.sidebarOpen" @click="close" />
  <aside class="sidebar" :class="{ open: ui.sidebarOpen }" aria-label="主导航">
    <div class="brand">
      <div class="brand-seal" aria-hidden="true">境</div>
      <div><strong>棋境</strong><span>XIANGQI LAB</span></div>
      <button class="icon-button sidebar-close" aria-label="关闭导航" @click="close">
        <AppIcon name="close" />
      </button>
    </div>
    <nav class="side-nav">
      <p class="nav-label">棋局</p>
      <RouterLink
        v-for="item in navigation"
        :key="item.to"
        class="nav-item"
        :class="{ active: route.path === item.to }"
        :to="item.to"
        @click="close"
      >
        <AppIcon :name="item.icon" />
        <span>{{ item.label }}</span>
        <span v-if="item.live" class="nav-dot" aria-label="有进行中的对局" />
      </RouterLink>
      <p class="nav-label">研习</p>
      <RouterLink
        v-for="item in learningNavigation"
        :key="item.to"
        class="nav-item"
        :class="{ active: route.path === item.to }"
        :to="item.to"
        @click="close"
      >
        <AppIcon :name="item.icon" />
        <span>{{ item.label }}</span>
        <span v-if="item.to === '/records' && recordCount !== null" class="nav-count">
          {{ recordCount }}
        </span>
      </RouterLink>
    </nav>
    <div class="sidebar-bottom">
      <RouterLink
        class="nav-item"
        :class="{ active: route.path === '/settings' }"
        to="/settings"
        @click="close"
      >
        <AppIcon name="settings" />
        <span>设置与诊断</span>
      </RouterLink>
      <div class="engine-status">
        <span class="status-light" :class="{ unavailable: engineUnavailable }" />
        <div>
          <strong>{{ engineHealth?.name || '引擎服务' }}</strong>
          <small>{{ engineStatusLabel }}</small>
        </div>
      </div>
    </div>
  </aside>
</template>
