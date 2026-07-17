<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import { useUiStore } from '@/stores/ui'

const route = useRoute()
const ui = useUiStore()
const title = computed(() => String(route.meta.title ?? '棋境'))
const eyebrow = computed(() => String(route.meta.eyebrow ?? 'Xiangqi Lab'))
const showStartAction = computed(() => !route.path.startsWith('/match/'))
</script>

<template>
  <header class="topbar">
    <button class="icon-button mobile-menu" aria-label="打开导航" @click="ui.sidebarOpen = true"><AppIcon name="menu" /></button>
    <div class="page-heading"><span>{{ eyebrow }}</span><h1>{{ title }}</h1></div>
    <div class="top-actions">
      <button class="icon-button" aria-label="切换深浅主题" @click="ui.toggleTheme"><AppIcon :name="ui.resolvedTheme === 'dark' ? 'sun' : 'moon'" /></button>
      <RouterLink v-if="showStartAction" class="primary-button compact" to="/new-game"><AppIcon name="play" />开始对局</RouterLink>
    </div>
  </header>
</template>
