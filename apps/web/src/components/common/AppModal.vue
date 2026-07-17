<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, useId, watch } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'

const props = withDefaults(defineProps<{
  open: boolean
  title: string
  description: string
  danger?: boolean
  confirmLabel?: string
  cancelLabel?: string
}>(), {
  danger: false,
  confirmLabel: '',
  cancelLabel: '取消',
})

const emit = defineEmits<{ close: []; confirm: [] }>()
const dialog = ref<HTMLElement | null>(null)
const titleId = `dialog-title-${useId()}`
const descriptionId = `dialog-description-${useId()}`
let previousFocus: HTMLElement | null = null

const focusableSelector = [
  'button:not([disabled])',
  '[href]',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',')

function close() {
  emit('close')
}

function handleKeydown(event: KeyboardEvent) {
  if (!props.open) return
  if (event.key === 'Escape') {
    event.preventDefault()
    close()
    return
  }
  if (event.key !== 'Tab' || !dialog.value) return

  const focusable = [...dialog.value.querySelectorAll<HTMLElement>(focusableSelector)]
  if (!focusable.length) {
    event.preventDefault()
    dialog.value.focus()
    return
  }
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last?.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first?.focus()
  }
}

watch(() => props.open, async (open) => {
  if (open) {
    previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
    document.addEventListener('keydown', handleKeydown)
    await nextTick()
    dialog.value?.querySelector<HTMLElement>(focusableSelector)?.focus()
    return
  }
  document.removeEventListener('keydown', handleKeydown)
  previousFocus?.focus()
  previousFocus = null
}, { immediate: true })

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
  previousFocus?.focus()
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="open" class="modal-overlay" role="presentation" @click.self="close">
        <section
          ref="dialog"
          class="modal-card"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="titleId"
          :aria-describedby="descriptionId"
          tabindex="-1"
        >
          <button class="icon-button dialog-close" aria-label="关闭" @click="close"><AppIcon name="close" /></button>
          <span class="dialog-icon" :class="{ green: !danger }"><AppIcon :name="danger ? 'flag' : 'spark'" /></span>
          <h2 :id="titleId">{{ title }}</h2>
          <p :id="descriptionId">{{ description }}</p>
          <slot />
          <div v-if="danger || confirmLabel" class="modal-actions">
            <button class="secondary-button" @click="close">{{ cancelLabel }}</button>
            <button :class="danger ? 'danger-button' : 'primary-button'" @click="emit('confirm')">
              {{ confirmLabel || '确认认输' }}
            </button>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>
