import { ref } from 'vue'
import { defineStore } from 'pinia'
import {
  deleteRecord as requestDeleteRecord,
  getRecord,
  importRecordFile,
  listRecords,
} from '@/api/records'
import type { GameRecord } from '@/api/contracts'

export const useRecordsStore = defineStore('records', () => {
  const records = ref<GameRecord[]>([])
  const importing = ref(false)
  const importProgress = ref(0)
  const importTitle = ref('')
  const totalRecords = ref(0)
  const loaded = ref(false)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchRecords() {
    loading.value = true
    error.value = null
    try {
      const items = await listRecords()
      records.value = items
      totalRecords.value = items.length
      loaded.value = true
    } catch {
      error.value = '无法从后端获取棋谱列表'
    } finally {
      loading.value = false
    }
  }

  async function importRecords(files: File[], collectionName?: string): Promise<number> {
    if (files.length === 0) return 0
    importing.value = true
    importProgress.value = 0
    importTitle.value = '正在导入…'
    error.value = null

    let imported = 0
    let failed = 0
    for (let i = 0; i < files.length; i++) {
      const file = files[i]!
      importTitle.value = `正在处理：${file.name}`

      try {
        const batch = await importRecordFile(file, collectionName)
        imported += batch.importedGames
        failed += batch.failedGames
        importProgress.value = Math.round(((i + 1) / files.length) * 100)
      } catch {
        failed += 1
      }
    }

    importTitle.value = failed > 0
      ? `导入 ${imported} 盘，失败 ${failed} 个文件`
      : `${imported} 盘棋谱已导入`
    importing.value = false
    importProgress.value = 100

    await fetchRecords()
    return imported
  }

  async function deleteRecord(id: string) {
    await requestDeleteRecord(id)
    records.value = records.value.filter((r) => r.id !== id)
    totalRecords.value = records.value.length
  }

  async function fetchRecord(id: string): Promise<GameRecord | null> {
    try {
      return await getRecord(id)
    } catch {
      return null
    }
  }

  function reset() {
    records.value = []
    importing.value = false
    importProgress.value = 0
    importTitle.value = ''
    totalRecords.value = 0
    loaded.value = false
    loading.value = false
    error.value = null
  }

  return {
    records, importing, importProgress, importTitle, totalRecords, loaded, loading, error,
    fetchRecords, importRecords, deleteRecord, fetchRecord, reset,
  }
})
