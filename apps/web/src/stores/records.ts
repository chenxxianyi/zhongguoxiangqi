import { ref } from 'vue'
import { defineStore } from 'pinia'
import { apiRequest } from '@/api/client'
import type { GameRecord, ImportBatch } from '@/api/contracts'

export const useRecordsStore = defineStore('records', () => {
  const records = ref<GameRecord[]>([])
  const importing = ref(false)
  const importProgress = ref(0)
  const importTitle = ref('')
  const totalRecords = ref(0)
  const loaded = ref(false)

  // ── 获取棋谱列表 ──
  async function fetchRecords() {
    try {
      const result = await apiRequest<{ items: GameRecord[] }>('/records')
      records.value = result.items
      totalRecords.value = result.items.length
      loaded.value = true
    } catch {
      console.warn('无法获取棋谱列表')
    }
  }

  // ── 导入棋谱 ──
  async function importRecords(files: File[], collectionName?: string) {
    if (files.length === 0) return
    importing.value = true
    importProgress.value = 0
    importTitle.value = '正在导入…'

    let imported = 0
    for (let i = 0; i < files.length; i++) {
      const file = files[i]!
      importTitle.value = `正在处理：${file.name}`

      try {
        const formData = new FormData()
        formData.append('file', file)
        if (collectionName) formData.append('name', collectionName)
        if (file.name.endsWith('.pgn')) formData.append('format', 'pgn')
        else if (file.name.endsWith('.json')) formData.append('format', 'json')
        else formData.append('format', 'iccs')

        const batch = await apiRequest<ImportBatch>('/records/imports', {
          method: 'POST',
          headers: {}, // 让浏览器自动设置 multipart Content-Type
          body: formData,
        })
        imported += batch.importedGames
        importProgress.value = Math.round(((i + 1) / files.length) * 100)
      } catch {
        // 单个文件失败不影响后续
        console.warn(`导入失败：${file.name}`)
      }
    }

    importTitle.value = `${imported} 盘棋谱已导入`
    importing.value = false
    importProgress.value = 100

    // 刷新列表
    await fetchRecords()
    return imported
  }

  // ── 删除棋谱 ──
  async function deleteRecord(id: string) {
    await apiRequest(`/records/${id}`, { method: 'DELETE' })
    records.value = records.value.filter((r) => r.id !== id)
    totalRecords.value = records.value.length
  }

  // ── 获取单条棋谱 ──
  async function fetchRecord(id: string): Promise<GameRecord | null> {
    try {
      return await apiRequest<GameRecord>(`/records/${id}`)
    } catch {
      return null
    }
  }

  // ── 重置 ──
  function reset() {
    records.value = []
    importing.value = false
    importProgress.value = 0
    importTitle.value = ''
    totalRecords.value = 0
    loaded.value = false
  }

  return {
    records, importing, importProgress, importTitle, totalRecords, loaded,
    fetchRecords, importRecords, deleteRecord, fetchRecord, reset,
  }
})
