import { apiRequest } from './client'
import type { DifficultyProfile } from './contracts'

export interface EngineHealth {
  name: string
  status: string
  type: string
}

export interface ExternalEngineLicense {
  name: string
  status: string
  notice: string
}

export interface LicenseInfo {
  application: string
  externalEngines: ExternalEngineLicense[]
}

export async function listDifficultyProfiles(): Promise<DifficultyProfile[]> {
  const response = await apiRequest<{ items: DifficultyProfile[] }>('/difficulty-profiles')
  return response.items
}

export function getEngineHealth(): Promise<EngineHealth> {
  return apiRequest<EngineHealth>('/engines/health')
}

export function getLicenses(): Promise<LicenseInfo> {
  return apiRequest<LicenseInfo>('/about/licenses')
}
