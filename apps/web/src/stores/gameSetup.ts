import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { apiRequest } from '@/api/client'
import type { AiMode, SideChoice } from '@/types/xiangqi'
import type { DifficultyProfile } from '@/api/contracts'

export const useGameSetupStore = defineStore('gameSetup', () => {
  const side = ref<SideChoice>('red')
  const difficulty = ref(6)
  const mode = ref<AiMode>('standard')
  const profiles = ref<DifficultyProfile[]>([])
  const loaded = ref(false)

  // ── 获取难度配置 ──
  async function fetchProfiles() {
    try {
      const result = await apiRequest<{ items: DifficultyProfile[] }>('/difficulty-profiles')
      profiles.value = result.items
      loaded.value = true
    } catch {
      // 如果无法连接后端，使用静默降级
      console.warn('无法获取难度配置，使用默认值')
    }
  }

  // ── 当前选中的难度配置 ──
  const profile = computed<DifficultyProfile>(() => {
    if (profiles.value.length > 0) {
      return profiles.value[difficulty.value - 1] ?? profiles.value[5] ?? profiles.value[0]!
    }
    // 降级：返回静态默认
    return {
      level: difficulty.value,
      name: `${difficulty.value} 级`,
      moveTimeMs: 300 + difficulty.value * 100,
      maxDepth: Math.min(6, Math.ceil(difficulty.value / 2)),
      maxNodes: Math.pow(2, difficulty.value + 8),
      multiPV: 3,
      description: '难度自适应配置',
    }
  })

  const modeLabel = computed(() =>
    ({ standard: '标准引擎', library: '棋谱库优先', style: '棋风模仿' })[mode.value],
  )
  const sideLabel = computed(() =>
    ({ red: '红方', black: '黑方', random: '随机执色' })[side.value],
  )

  // ── 立即加载 ──
  fetchProfiles()

  return { side, difficulty, mode, profiles, loaded, profile, modeLabel, sideLabel, fetchProfiles }
})
