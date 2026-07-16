import { apiRequest } from './client'
import type { GameRecord, ImportBatch } from './contracts'

export async function listRecords(): Promise<GameRecord[]> {
  const response = await apiRequest<{ items: GameRecord[] }>('/records')
  return response.items
}

export function importRecordFile(file: File, collectionName?: string): Promise<ImportBatch> {
  const formData = new FormData()
  formData.append('file', file)
  if (collectionName) formData.append('name', collectionName)

  const fileName = file.name.toLowerCase()
  const format = fileName.endsWith('.pgn')
    ? 'pgn'
    : fileName.endsWith('.json')
      ? 'json'
      : 'iccs'
  formData.append('format', format)

  return apiRequest<ImportBatch>('/records/imports', {
    method: 'POST',
    body: formData,
  })
}

export function deleteRecord(id: string): Promise<void> {
  return apiRequest<void>(`/records/${encodeURIComponent(id)}`, { method: 'DELETE' })
}

export function getRecord(id: string): Promise<GameRecord> {
  return apiRequest<GameRecord>(`/records/${encodeURIComponent(id)}`)
}
