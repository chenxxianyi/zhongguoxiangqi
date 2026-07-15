import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export const useLearningStore = defineStore('learning', () => {
  const progress = ref(0)
  const running = ref(false)
  const completed = computed(() => progress.value >= 100)
  const stages = ['安全校验', '逐着规则验证', '局面着法统计', '棋风特征提取', '质量检查', '构建完成']
  const stage = computed(() => stages[Math.min(stages.length - 1, Math.floor(progress.value / 20))] ?? stages[0])
  let timer: number | undefined

  function start() {
    if (running.value) return
    progress.value = 0
    running.value = true
    timer = window.setInterval(() => {
      progress.value = Math.min(100, progress.value + 4)
      if (progress.value >= 100) {
        running.value = false
        window.clearInterval(timer)
      }
    }, 110)
  }

  return { progress, running, completed, stage, start }
})
