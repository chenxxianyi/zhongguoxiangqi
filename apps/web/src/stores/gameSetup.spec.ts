import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useGameSetupStore } from './gameSetup'

const listDifficultyProfiles = vi.hoisted(() => vi.fn())

vi.mock('@/api/system', () => ({ listDifficultyProfiles }))

describe('game setup store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    listDifficultyProfiles.mockReset()
  })

  it('uses only difficulty profiles returned by the backend', async () => {
    listDifficultyProfiles.mockResolvedValue([
      {
        level: 2,
        name: '后端档位',
        moveTimeMs: 120,
        maxDepth: 2,
        maxNodes: 2000,
        multiPV: 3,
        description: '后端配置',
      },
    ])

    const store = useGameSetupStore()
    await store.fetchProfiles()

    expect(store.profile?.name).toBe('后端档位')
    expect(store.difficulty).toBe(2)
    expect(store.error).toBeNull()
  })

  it('does not synthesize a profile when the backend is unavailable', async () => {
    listDifficultyProfiles.mockRejectedValue(new Error('offline'))

    const store = useGameSetupStore()
    await store.fetchProfiles()

    expect(store.profile).toBeNull()
    expect(store.profiles).toEqual([])
    expect(store.error).toBe('无法从后端获取难度配置')
  })
})
