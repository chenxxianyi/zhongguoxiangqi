<script setup lang="ts">
import { RouterView } from 'vue-router'
import IconSprite from '@/components/common/IconSprite.vue'
import ToastRegion from '@/components/common/ToastRegion.vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
</script>

<template>
  <a class="skip-link" href="#main-content">跳到主要内容</a>
  <IconSprite />
  <div class="app-shell" :class="{ 'focus-mode': ui.matchFocusMode }">
    <AppSidebar />
    <div class="main-column">
      <AppHeader />
      <main id="main-content" tabindex="-1">
        <RouterView v-slot="{ Component, route }">
          <Transition name="page-route" mode="out-in">
            <component :is="Component" :key="route.fullPath" />
          </Transition>
        </RouterView>
      </main>
    </div>
  </div>
  <ToastRegion />
</template>
