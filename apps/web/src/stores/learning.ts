import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { apiRequest } from '@/api/client'
import type { LearningJob, LearningVersion } from '@/api/contracts'

export const useLearningStore = defineStore('learning', () => {
  const progress = ref(0)
  const running = ref(false)
  const currentJob = ref<LearningJob | null>(null)
  const versions = ref<LearningVersion[]>([])
  const activeVersion = ref<LearningVersion | null>(null)
  const loaded = ref(false)
  const stages = ref<string[]>([])
  const stage = computed(() => {
    if (currentJob.value?.status === 'running') {
      const idx = Math.min(stages.value.length - 1, Math.floor(progress.value / 25))
      return stages.value[idx] ?? '处理中…'
    }
    if (currentJob.value?.status === 'completed') return '构建完成'
    if (currentJob.value?.status === 'failed') return '构建失败'
    return '等待开始'
  })
  const completed = computed(() => currentJob.value?.status === 'completed')
  let pollTimer: number | undefined

  // ── 获取版本列表 ──
  async function fetchVersions() {
    try {
      const result = await apiRequest<{ items: LearningVersion[] }>('/learning/versions')
      versions.value = result.items
      // 找到当前启用的版本
      activeVersion.value = result.items.find((v) => v.status === 'active') ?? null
      loaded.value = true
    } catch {
      console.warn('无法获取学习版本列表')
    }
  }

  // ── 创建构建任务 ──
  async function createJob(name?: string, recordIds?: string[]) {
    if (running.value) return
    running.value = true
    progress.value = 0
    stages.value = ['安全校验', '逐着规则验证', '局面着法统计', '棋风特征提取', '质量检查', '构建完成']

    try {
      currentJob.value = await apiRequest<LearningJob>('/learning/jobs', {
        method: 'POST',
        body: JSON.stringify({ name: name || '棋谱学习版本', recordIds }),
      })

      // 轮询构建进度
      await pollJob(currentJob.value.id)
    } catch {
      currentJob.value = {
        id: '', status: 'failed', name: '构建失败',
        progress: 0, recordCount: 0, moveCount: 0,
        createdAt: new Date().toISOString(),
      }
      running.value = false
    }
  }

  // ── 轮询任务状态 ──
  async function pollJob(jobId: string) {
    return new Promise<void>((resolve) => {
      const poll = async () => {
        try {
          const job = await apiRequest<LearningJob>(`/learning/jobs/${jobId}`)
          currentJob.value = job
          progress.value = job.progress

          if (job.status === 'completed' || job.status === 'failed') {
            running.value = false
            if (pollTimer) clearInterval(pollTimer)
            await fetchVersions()
            resolve()
            return
          }
          pollTimer = window.setTimeout(poll, 500)
        } catch {
          running.value = false
          if (pollTimer) clearInterval(pollTimer)
          resolve()
        }
      }
      pollTimer = window.setTimeout(poll, 200)
    })
  }

  // ── 激活版本 ──
  async function activateVersion(id: string) {
    try {
      const version = await apiRequest<LearningVersion>(`/learning/versions/${id}/activate`, {
        method: 'POST',
      })
      activeVersion.value = version
      await fetchVersions()
      return true
    } catch {
      return false
    }
  }

  // ── 回滚版本 ──
  async function rollback(id: string) {
    try {
      const version = await apiRequest<LearningVersion>(`/learning/versions/${id}/rollback`, {
        method: 'POST',
      })
      activeVersion.value = version
      await fetchVersions()
      return true
    } catch {
      return false
    }
  }

  // ── 重置 ──
  function reset() {
    if (pollTimer) clearInterval(pollTimer)
    progress.value = 0
    running.value = false
    currentJob.value = null
  }

  return {
    progress, running, completed, stage, stages,
    versions, activeVersion, loaded, currentJob,
    fetchVersions, createJob, activateVersion, rollback, reset,
  }
})
