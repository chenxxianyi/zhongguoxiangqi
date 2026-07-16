import { apiRequest } from './client'
import type { LearningJob, LearningVersion } from './contracts'

export interface CreateLearningJobRequest {
  name: string
  recordIds?: string[]
}

export async function listLearningVersions(): Promise<LearningVersion[]> {
  const response = await apiRequest<{ items: LearningVersion[] }>('/learning/versions')
  return response.items
}

export function createLearningJob(request: CreateLearningJobRequest): Promise<LearningJob> {
  return apiRequest<LearningJob>('/learning/jobs', {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

export function getLearningJob(id: string): Promise<LearningJob> {
  return apiRequest<LearningJob>(`/learning/jobs/${encodeURIComponent(id)}`)
}

export function activateLearningVersion(id: string): Promise<LearningVersion> {
  return apiRequest<LearningVersion>(`/learning/versions/${encodeURIComponent(id)}/activate`, {
    method: 'POST',
  })
}

export function rollbackLearningVersion(id: string): Promise<LearningVersion> {
  return apiRequest<LearningVersion>(`/learning/versions/${encodeURIComponent(id)}/rollback`, {
    method: 'POST',
  })
}
