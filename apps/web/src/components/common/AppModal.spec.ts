import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import AppModal from './AppModal.vue'

describe('AppModal', () => {
  it('closes with Escape and restores the previous focus', async () => {
    const trigger = document.createElement('button')
    document.body.append(trigger)
    trigger.focus()

    const wrapper = mount(AppModal, {
      attachTo: document.body,
      props: {
        open: true,
        title: '确认操作',
        description: '操作说明',
      },
    })
    await wrapper.vm.$nextTick()

    expect(document.activeElement?.getAttribute('aria-label')).toBe('关闭')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    expect(wrapper.emitted('close')).toHaveLength(1)

    await wrapper.setProps({ open: false })
    expect(document.activeElement).toBe(trigger)
    wrapper.unmount()
    trigger.remove()
  })

  it('wraps keyboard focus inside the dialog', async () => {
    const wrapper = mount(AppModal, {
      attachTo: document.body,
      props: {
        open: true,
        title: '删除棋谱',
        description: '不可撤销',
        danger: true,
        confirmLabel: '确认删除',
      },
    })
    await wrapper.vm.$nextTick()

    const buttons = [...document.querySelectorAll<HTMLElement>('.modal-card button')]
    buttons.at(-1)?.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab' }))
    expect(document.activeElement).toBe(buttons[0])
    wrapper.unmount()
  })
})
