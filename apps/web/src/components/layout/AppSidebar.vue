<script setup lang="ts">
import { RouterLink, useRoute } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import { useUiStore } from '@/stores/ui'

const route = useRoute()
const ui = useUiStore()

const navigation = [
  { label: '首页', icon: 'grid', to: '/' },
  { label: '新对局', icon: 'play', to: '/new-game' },
  { label: '对弈棋盘', icon: 'board', to: '/match', live: true },
  { label: '历史对局', icon: 'clock', to: '/history' },
]
const learningNavigation = [
  { label: '棋谱库', icon: 'record', to: '/records', count: '128' },
  { label: '学习中心', icon: 'spark', to: '/learning' },
  { label: '复盘分析', icon: 'chart', to: '/analysis' },
]

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
      <button class="icon-button sidebar-close" aria-label="关闭导航" @click="close"><AppIcon name="close" /></button>
    </div>
    <nav class="side-nav">
      <p class="nav-label">棋局</p>
      <RouterLink v-for="item in navigation" :key="item.to" class="nav-item" :class="{ active: route.path === item.to }" :to="item.to" @click="close">
        <AppIcon :name="item.icon" /><span>{{ item.label }}</span><span v-if="item.live" class="nav-dot" aria-label="有进行中的对局" />
      </RouterLink>
      <p class="nav-label">研习</p>
      <RouterLink v-for="item in learningNavigation" :key="item.to" class="nav-item" :class="{ active: route.path === item.to }" :to="item.to" @click="close">
        <AppIcon :name="item.icon" /><span>{{ item.label }}</span><span v-if="item.count" class="nav-count">{{ item.count }}</span>
      </RouterLink>
    </nav>
    <div class="sidebar-bottom">
      <RouterLink class="nav-item" :class="{ active: route.path === '/settings' }" to="/settings" @click="close"><AppIcon name="settings" /><span>设置与诊断</span></RouterLink>
      <div class="engine-status"><span class="status-light" /><div><strong>演示引擎状态</strong><small>未连接真实后端</small></div></div>
      <div class="profile-mini"><div class="avatar">林</div><div><strong>林间棋客</strong><span>进阶棋手</span></div><button class="icon-button" aria-label="账户选项"><AppIcon name="more" /></button></div>
    </div>
  </aside>
</template>
