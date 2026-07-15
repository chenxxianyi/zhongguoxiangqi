<script setup lang="ts">
import { useUiStore } from '@/stores/ui'
const ui = useUiStore()
</script>

<template>
  <div class="toast-region" aria-live="polite" aria-atomic="true">
    <button v-for="toast in ui.toasts" :key="toast.id" class="toast" @click="ui.dismissToast(toast.id)">{{ toast.message }}</button>
  </div>
</template>
