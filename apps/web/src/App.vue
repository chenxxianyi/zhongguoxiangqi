<script setup lang="ts">
import { RouterView } from 'vue-router'
import IconSprite from '@/components/common/IconSprite.vue'
import ToastRegion from '@/components/common/ToastRegion.vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
</script>

<template>
  <a class="skip-link" href="#main-content">跳到主要内容</a>
  <IconSprite />
  <div class="app-shell">
    <AppSidebar />
    <div class="main-column">
      <AppHeader />
      <main id="main-content" tabindex="-1"><RouterView /></main>
    </div>
  </div>
  <ToastRegion />
</template>
