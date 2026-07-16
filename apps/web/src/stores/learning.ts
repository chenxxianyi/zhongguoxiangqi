import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  activateLearningVersion,
  createLearningJob,
  getLearningJob,
  listLearningVersions,
  rollbackLearningVersion,
} from '@/api/learning'
import type { LearningJob, LearningVersion } from '@/api/contracts'

export const useLearningStore = defineStore('learning', () => {
  const progress = ref(0)
  const running = ref(false)
  const currentJob = ref<LearningJob | null>(null)
  const versions = ref<LearningVersion[]>([])
  const activeVersion = ref<LearningVersion | null>(null)
  const loaded = ref(false)
  const versionsLoading = ref(false)
  const versionsError = ref<string | null>(null)
  const jobError = ref<string | null>(null)
  const stage = computed(() => {
    if (currentJob.value?.message) return currentJob.value.message
    if (currentJob.value?.status === 'queued') return '等待后端处理'
    if (currentJob.value?.status === 'running') return '后端正在构建'
    if (currentJob.value?.status === 'completed') return '构建完成'
    if (currentJob.value?.status === 'failed') return '构建失败'
    return '等待开始'
  })
  const completed = computed(() => currentJob.value?.status === 'completed')
  let pollTimer: number | undefined

  async function fetchVersions() {
    versionsLoading.value = true
    versionsError.value = null
    try {
      const items = await listLearningVersions()
      versions.value = items
      activeVersion.value = items.find((version) => version.status === 'active') ?? null
      loaded.value = true
    } catch {
      versionsError.value = '无法从后端获取学习版本'
    } finally {
      versionsLoading.value = false
    }
  }

  async function createJob(name?: string, recordIds?: string[]) {
    if (running.value) return
    running.value = true
    progress.value = 0
    jobError.value = null
    currentJob.value = null

    try {
      currentJob.value = await createLearningJob({
        name: name?.trim() ?? '',
        recordIds,
      })
      await pollJob(currentJob.value.id)
    } catch {
      jobError.value = '创建学习任务失败'
      running.value = false
    }
  }

  async function pollJob(jobId: string) {
    return new Promise<void>((resolve) => {
      const poll = async () => {
        try {
          const job = await getLearningJob(jobId)
          currentJob.value = job
          progress.value = job.progress

          if (job.status === 'completed') {
            running.value = false
            if (pollTimer) clearTimeout(pollTimer)
            await fetchVersions()
            resolve()
            return
          }
          if (job.status === 'failed') {
            running.value = false
            jobError.value = job.message || '学习版本构建失败'
            if (pollTimer) clearTimeout(pollTimer)
            resolve()
            return
          }
          pollTimer = window.setTimeout(poll, 500)
        } catch {
          running.value = false
          jobError.value = '无法获取学习任务状态'
          if (pollTimer) clearTimeout(pollTimer)
          resolve()
        }
      }
      pollTimer = window.setTimeout(poll, 200)
    })
  }

  async function activateVersion(id: string) {
    try {
      activeVersion.value = await activateLearningVersion(id)
      await fetchVersions()
      return true
    } catch {
      versionsError.value = '启用学习版本失败'
      return false
    }
  }

  async function rollback(id: string) {
    try {
      activeVersion.value = await rollbackLearningVersion(id)
      await fetchVersions()
      return true
    } catch {
      versionsError.value = '回滚学习版本失败'
      return false
    }
  }

  function reset() {
    if (pollTimer) clearTimeout(pollTimer)
    progress.value = 0
    running.value = false
    currentJob.value = null
    jobError.value = null
  }

  return {
    progress,
    running,
    completed,
    stage,
    versions,
    activeVersion,
    loaded,
    versionsLoading,
    versionsError,
    jobError,
    currentJob,
    fetchVersions,
    createJob,
    activateVersion,
    rollback,
    reset,
  }
})
