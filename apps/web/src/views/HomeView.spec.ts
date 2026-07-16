import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import HomeView from './HomeView.vue'

const {
  listMatches,
  listRecords,
  listLearningVersions,
} = vi.hoisted(() => ({
  listMatches: vi.fn(),
  listRecords: vi.fn(),
  listLearningVersions: vi.fn(),
}))

vi.mock('@/api/matches', () => ({ listMatches }))
vi.mock('@/api/records', () => ({ listRecords }))
vi.mock('@/api/learning', () => ({ listLearningVersions }))

describe('HomeView learning summary', () => {
  beforeEach(() => {
    listMatches.mockResolvedValue([])
    listRecords.mockResolvedValue([{ id: 'record-1' }, { id: 'record-2' }])
    listLearningVersions.mockResolvedValue([
      { id: 'version-1', status: 'active' },
      { id: 'version-2', status: 'ready' },
    ])
  })

  it('renders record and learning version counts from backend data', async () => {
    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          MiniBoard: true,
          AppIcon: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('2 盘')
    expect(wrapper.text()).toContain('学习版本2')
    expect(wrapper.text()).toContain('已启用')
    expect(wrapper.text()).toContain('查看学习版本')
  })

  it('shows an unavailable state instead of presenting failed requests as empty data', async () => {
    listRecords.mockRejectedValue(new Error('offline'))

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          MiniBoard: true,
          AppIcon: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('服务不可用')
  })
})
