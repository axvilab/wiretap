import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ReplayModal from '@/components/ReplayModal.vue'

vi.mock('@/api', () => ({
  api: {
    replayRequest: vi.fn().mockResolvedValue({ success: true, status: 200 }),
    retryJob: vi.fn().mockResolvedValue({ success: true, output: 'retried' }),
    dispatchJob: vi.fn().mockResolvedValue({ success: true, job_class: 'TestJob' }),
    replayBroadcast: vi.fn().mockResolvedValue({ success: true }),
    replayCommand: vi.fn().mockResolvedValue({ success: true, exit_code: 0, output: 'done' }),
  },
}))

const globalStubs = { teleport: true }

async function mountModal(type: string, data: Record<string, any>) {
  const wrapper = mount(ReplayModal, {
    props: { show: false, type: type as any, data },
    global: { stubs: globalStubs },
  })
  await wrapper.setProps({ show: true })
  await wrapper.vm.$nextTick()
  return wrapper
}

describe('ReplayModal', () => {
  it('does not render content when show is false', () => {
    const wrapper = mount(ReplayModal, {
      props: { show: false, type: 'request', data: {} },
      global: { stubs: globalStubs },
    })
    expect(wrapper.find('.fixed').exists()).toBe(false)
  })

  it('renders request form', async () => {
    const wrapper = await mountModal('request', { method: 'POST', uri: '/api/test', body: '{"name":"test"}' })
    expect(wrapper.text()).toContain('Replay Request')
    expect(wrapper.find('select').element.value).toBe('POST')
  })

  it('renders job form with payload', async () => {
    const wrapper = await mountModal('job', {
      job_class: 'App\\Jobs\\Test',
      job_serialized: 'base64data',
      job_payload: { orderId: 42 },
      status: 'completed',
    })
    expect(wrapper.text()).toContain('Replay Job')
    expect(wrapper.text()).toContain('Properties (JSON)')
    const classInput = wrapper.find('input[readonly]')
    expect((classInput.element as HTMLInputElement).value).toBe('App\\Jobs\\Test')
  })

  it('renders job retry form for failed without payload', async () => {
    const wrapper = await mountModal('job', { uuid: 'abc-123', status: 'failed' })
    expect(wrapper.text()).toContain('Retry Job')
    expect(wrapper.text()).toContain('queue:retry')
  })

  it('renders broadcast form', async () => {
    const wrapper = await mountModal('broadcast', {
      event: 'OrderShipped',
      channels: [{ name: 'orders', type: 'public' }],
      payload: { id: 1 },
    })
    expect(wrapper.text()).toContain('Replay Broadcast')
    expect(wrapper.text()).toContain('Event Name')
  })

  it('renders command form', async () => {
    const wrapper = await mountModal('command', { command: 'inspire', input: '' })
    expect(wrapper.text()).toContain('Replay Command')
  })

  it('shows masked fields warning', async () => {
    const wrapper = await mountModal('request', { method: 'POST', uri: '/login', body: '{"password":"***"}' })
    expect(wrapper.text()).toContain('masked fields')
  })

  it('shows large payload warning', async () => {
    const wrapper = await mountModal('job', {
      job_class: 'BigJob',
      job_serialized: 'data',
      job_payload: {},
      payload_size: 100000,
      status: 'completed',
    })
    expect(wrapper.text()).toContain('Large job payload')
  })

  it('emits close on cancel', async () => {
    const wrapper = await mountModal('request', {})
    const cancelBtn = wrapper.findAll('button').find(b => b.text() === 'Cancel')
    expect(cancelBtn).toBeTruthy()
    await cancelBtn!.trigger('click')
    expect(wrapper.emitted('close')).toBeTruthy()
  })
})
