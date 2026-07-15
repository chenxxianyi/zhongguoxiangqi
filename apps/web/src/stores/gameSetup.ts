import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { difficultyProfiles } from '@/data/demo'
import type { AiMode, SideChoice } from '@/types/xiangqi'

export const useGameSetupStore = defineStore('gameSetup', () => {
  const side = ref<SideChoice>('red')
  const difficulty = ref(6)
  const mode = ref<AiMode>('standard')

  const profile = computed(() => difficultyProfiles[difficulty.value - 1] ?? difficultyProfiles[5]!)
  const modeLabel = computed(() => ({ standard: '标准引擎', library: '棋谱库优先', style: '棋风模仿' })[mode.value])
  const sideLabel = computed(() => ({ red: '红方', black: '黑方', random: '随机执色' })[side.value])

  return { side, difficulty, mode, profile, modeLabel, sideLabel }
})
