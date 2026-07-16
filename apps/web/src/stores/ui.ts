import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'

export type ThemeChoice = 'light' | 'dark' | 'system'

export const useUiStore = defineStore('ui', () => {
  const sidebarOpen = ref(false)
  const saved = localStorage.getItem('xiangqi-theme') as ThemeChoice | null
  const theme = ref<ThemeChoice>(saved ?? 'light')
  const toasts = ref<Array<{ id: number; message: string }>>([])
  let toastId = 0

  const colorSchemeMedia = window.matchMedia?.('(prefers-color-scheme: dark)')
  const prefersDark = () => colorSchemeMedia?.matches ?? false
  const resolvedTheme = computed(() => theme.value === 'system' ? (prefersDark() ? 'dark' : 'light') : theme.value)

  function applyTheme() {
    document.documentElement.dataset.theme = resolvedTheme.value
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', resolvedTheme.value === 'dark' ? '#141916' : '#f3eee4')
  }

  function setTheme(value: ThemeChoice) {
    theme.value = value
  }

  function toggleTheme() {
    setTheme(resolvedTheme.value === 'dark' ? 'light' : 'dark')
  }

  function showToast(message: string) {
    const id = ++toastId
    toasts.value.push({ id, message })
    window.setTimeout(() => dismissToast(id), 3200)
  }

  function dismissToast(id: number) {
    toasts.value = toasts.value.filter((toast) => toast.id !== id)
  }

  watch(theme, (value) => {
    localStorage.setItem('xiangqi-theme', value)
    applyTheme()
  }, { immediate: true })

  colorSchemeMedia?.addEventListener?.('change', () => {
    if (theme.value === 'system') applyTheme()
  })

  return { sidebarOpen, theme, resolvedTheme, toasts, setTheme, toggleTheme, showToast, dismissToast }
})
