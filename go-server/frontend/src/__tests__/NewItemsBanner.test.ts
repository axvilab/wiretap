import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import NewItemsBanner from '@/components/NewItemsBanner.vue'

// Mock useWebSocket — capture the handler so we can call it in tests
let wsHandler: ((event: any) => void) | null = null
vi.mock('@/hooks/useWebSocket', () => ({
  useWebSocket: (handler: (event: any) => void) => {
    wsHandler = handler
    return { connected: { value: true } }
  },
}))

function simulateTrace(type: string, extras: Record<string, any> = {}) {
  wsHandler?.({
    type: 'trace.new',
    data: { type, events_json: '[]', has_exception: false, ...extras },
  })
}

describe('NewItemsBanner', () => {
  it('renders nothing initially', () => {
    const wrapper = mount(NewItemsBanner)
    expect(wrapper.text()).toBe('')
  })

  it('shows banner after receiving matching events', async () => {
    const wrapper = mount(NewItemsBanner, { props: { types: ['job'] } })

    simulateTrace('job')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('1 new item available')

    simulateTrace('job')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('2 new items available')
  })

  it('ignores non-matching trace types', async () => {
    const wrapper = mount(NewItemsBanner, { props: { types: ['job'] } })

    simulateTrace('request')
    simulateTrace('command')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toBe('')
  })

  it('counts all types when types prop is empty', async () => {
    const wrapper = mount(NewItemsBanner)

    simulateTrace('request')
    simulateTrace('job')
    simulateTrace('command')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('3 new items available')
  })

  it('filters by exceptionsOnly', async () => {
    const wrapper = mount(NewItemsBanner, { props: { exceptionsOnly: true } })

    simulateTrace('request', { has_exception: false })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toBe('')

    simulateTrace('request', { has_exception: true })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('1 new item available')
  })

  it('filters by mailOnly', async () => {
    const wrapper = mount(NewItemsBanner, { props: { mailOnly: true } })

    simulateTrace('request', { events_json: JSON.stringify([{ type: 'query' }]) })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toBe('')

    simulateTrace('request', { events_json: JSON.stringify([{ type: 'mail.sent' }]) })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('1 new item available')
  })

  it('filters by broadcastOnly', async () => {
    const wrapper = mount(NewItemsBanner, { props: { broadcastOnly: true } })

    simulateTrace('request', { events_json: JSON.stringify([{ type: 'broadcast' }]) })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('1 new item available')
  })

  it('filters by logsOnly', async () => {
    const wrapper = mount(NewItemsBanner, { props: { logsOnly: true } })

    simulateTrace('request', { events_json: JSON.stringify([{ type: 'query' }]) })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toBe('')

    simulateTrace('request', { events_json: JSON.stringify([{ type: 'log' }]) })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('1 new item available')
  })

  it('emits refresh and resets on click', async () => {
    const wrapper = mount(NewItemsBanner, { props: { types: ['job'] } })

    simulateTrace('job')
    simulateTrace('job')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('2 new items')

    await wrapper.find('[class*="cursor-pointer"]').trigger('click')
    expect(wrapper.emitted('refresh')).toHaveLength(1)
    expect(wrapper.text()).toBe('')
  })

  it('ignores non trace.new events', async () => {
    const wrapper = mount(NewItemsBanner)

    wsHandler?.({ type: 'other.event', data: { type: 'request' } })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toBe('')
  })
})
