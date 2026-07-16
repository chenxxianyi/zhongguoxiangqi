import { ref } from 'vue'
import { defineStore } from 'pinia'
import { ApiError } from '@/api/client'
import { createAnalysisJob, getAnalysisJob, getMatchAnalysis } from '@/api/analysis'
import type { AnalysisJob, AnalysisResult } from '@/api/contracts'

export const useAnalysisStore = defineStore('analysis', () => {
  const currentResult = ref<AnalysisResult | null>(null)
  const currentJob = ref<AnalysisJob | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  let pollTimer: number | undefined

  // ── 获取分析结果（如果已存在）或创建分析任务 ──
  async function loadOrCreateAnalysis(matchId: string) {
    loading.value = true
    error.value = null

    // 先尝试获取已有结果
    try {
      const result = await getMatchAnalysis(matchId)
      currentResult.value = result
      currentJob.value = null
      loading.value = false
      return
    } catch (requestError) {
      if (!(requestError instanceof ApiError) || requestError.status !== 404) {
        error.value = '无法从后端获取分析结果'
        loading.value = false
        return
      }
    }

    // 创建分析任务
    try {
      const job = await createAnalysisJob(matchId)
      currentJob.value = job
      await pollJob(job.id, matchId)
    } catch {
      error.value = '创建分析任务失败'
      loading.value = false
    }
  }

  // ── 轮询分析任务 ──
  async function pollJob(jobId: string, matchId: string) {
    return new Promise<void>((resolve) => {
      const poll = async () => {
        try {
          const job = await getAnalysisJob(jobId)
          currentJob.value = job

          if (job.status === 'completed') {
            // 任务完成，获取结果
            const result = await getMatchAnalysis(matchId)
            currentResult.value = result
            loading.value = false
            if (pollTimer) clearTimeout(pollTimer)
            resolve()
            return
          }

          if (job.status === 'failed') {
            error.value = job.message || '分析失败'
            loading.value = false
            if (pollTimer) clearTimeout(pollTimer)
            resolve()
            return
          }

          pollTimer = window.setTimeout(poll, 500)
        } catch {
          error.value = '获取分析状态失败'
          loading.value = false
          if (pollTimer) clearTimeout(pollTimer)
          resolve()
        }
      }
      pollTimer = window.setTimeout(poll, 200)
    })
  }

  // ── 重置 ──
  function reset() {
    if (pollTimer) clearTimeout(pollTimer)
    currentResult.value = null
    currentJob.value = null
    loading.value = false
    error.value = null
  }

  return {
    currentResult, currentJob, loading, error,
    loadOrCreateAnalysis, reset,
  }
})
