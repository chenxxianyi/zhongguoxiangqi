import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { listDifficultyProfiles } from '@/api/system'
import type { AiMode, SideChoice } from '@/types/xiangqi'
import type { DifficultyProfile } from '@/api/contracts'

export const useGameSetupStore = defineStore('gameSetup', () => {
  const side = ref<SideChoice>('red')
  const difficulty = ref(0)
  const mode = ref<AiMode>('standard')
  const profiles = ref<DifficultyProfile[]>([])
  const loaded = ref(false)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchProfiles() {
    loading.value = true
    error.value = null
    try {
      const items = await listDifficultyProfiles()
      if (items.length === 0) {
        throw new Error('后端未返回难度配置')
      }
      profiles.value = items
      if (!items.some((item) => item.level === difficulty.value)) {
        difficulty.value = items.find((item) => item.level === 6)?.level ?? items[0]!.level
      }
      loaded.value = true
    } catch {
      profiles.value = []
      loaded.value = false
      difficulty.value = 0
      error.value = '无法从后端获取难度配置'
    } finally {
      loading.value = false
    }
  }

  const profile = computed<DifficultyProfile | null>(() =>
    profiles.value.find((item) => item.level === difficulty.value) ?? null,
  )
  const minDifficulty = computed(() => profiles.value[0]?.level ?? 0)
  const maxDifficulty = computed(() => profiles.value.at(-1)?.level ?? 0)

  const modeLabel = computed(() =>
    ({ standard: '标准引擎', library: '棋谱库优先', style: '棋风模仿' })[mode.value],
  )
  const sideLabel = computed(() =>
    ({ red: '红方', black: '黑方', random: '随机执色' })[side.value],
  )

  void fetchProfiles()

  return {
    side,
    difficulty,
    mode,
    profiles,
    loaded,
    loading,
    error,
    profile,
    minDifficulty,
    maxDifficulty,
    modeLabel,
    sideLabel,
    fetchProfiles,
  }
})
