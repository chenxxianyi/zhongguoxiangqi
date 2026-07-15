<script setup lang="ts">
import AppIcon from '@/components/common/AppIcon.vue'

defineProps<{ open: boolean; title: string; description: string; danger?: boolean }>()
const emit = defineEmits<{ close: []; confirm: [] }>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="modal-overlay" role="presentation" @click.self="emit('close')">
      <section class="modal-card" role="dialog" aria-modal="true" :aria-label="title">
        <button class="icon-button dialog-close" aria-label="关闭" @click="emit('close')"><AppIcon name="close" /></button>
        <span class="dialog-icon" :class="{ green: !danger }"><AppIcon :name="danger ? 'flag' : 'spark'" /></span>
        <h2>{{ title }}</h2><p>{{ description }}</p><slot />
        <div v-if="danger" class="modal-actions"><button class="secondary-button" @click="emit('close')">继续对局</button><button class="danger-button" @click="emit('confirm')">确认认输</button></div>
      </section>
    </div>
  </Teleport>
</template>
