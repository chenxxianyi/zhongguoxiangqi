import { apiRequest } from './client'
import type { AnalysisJob, AnalysisResult } from './contracts'

export function getMatchAnalysis(matchId: string): Promise<AnalysisResult> {
  return apiRequest<AnalysisResult>(`/matches/${encodeURIComponent(matchId)}/analysis`)
}

export function createAnalysisJob(matchId: string): Promise<AnalysisJob> {
  return apiRequest<AnalysisJob>('/analysis/jobs', {
    method: 'POST',
    body: JSON.stringify({ matchId }),
  })
}

export function getAnalysisJob(id: string): Promise<AnalysisJob> {
  return apiRequest<AnalysisJob>(`/analysis/jobs/${encodeURIComponent(id)}`)
}
