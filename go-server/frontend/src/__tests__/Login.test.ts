import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Login from '@/pages/Login.vue'

const mockFetch = vi.fn()
globalThis.fetch = mockFetch

beforeEach(() => {
  mockFetch.mockReset()
})

function jsonResponse(data: unknown, status = 200) {
  return Promise.resolve({
    ok: status >= 200 && status < 400,
    status,
    json: () => Promise.resolve(data),
  })
}

describe('Login', () => {
  it('shows "Create your password" when setup=true', () => {
    const wrapper = mount(Login, { props: { setup: true } })
    expect(wrapper.text()).toContain('Create your password')
  })

  it('shows "Sign in to continue" when setup is not set', () => {
    const wrapper = mount(Login)
    expect(wrapper.text()).toContain('Sign in to continue')
  })

  it('shows confirm password field only in setup mode', () => {
    const loginWrapper = mount(Login)
    const setupWrapper = mount(Login, { props: { setup: true } })

    const loginInputs = loginWrapper.findAll('input[type="password"]')
    const setupInputs = setupWrapper.findAll('input[type="password"]')

    expect(loginInputs).toHaveLength(1)
    expect(setupInputs).toHaveLength(2)
  })

  it('shows button text "Sign In" in login mode', () => {
    const wrapper = mount(Login)
    expect(wrapper.find('button[type="submit"]').text()).toBe('Sign In')
  })

  it('shows button text "Create Password" in setup mode', () => {
    const wrapper = mount(Login, { props: { setup: true } })
    expect(wrapper.find('button[type="submit"]').text()).toBe('Create Password')
  })

  it('shows error for password shorter than 6 characters', async () => {
    const wrapper = mount(Login)

    await wrapper.find('input[type="password"]').setValue('abc')
    await wrapper.find('form').trigger('submit')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Password must be at least 6 characters')
    expect(mockFetch).not.toHaveBeenCalled()
  })

  it('shows error when passwords do not match in setup mode', async () => {
    const wrapper = mount(Login, { props: { setup: true } })
    const inputs = wrapper.findAll('input[type="password"]')

    await inputs[0].setValue('password123')
    await inputs[1].setValue('different456')
    await wrapper.find('form').trigger('submit')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Passwords do not match')
    expect(mockFetch).not.toHaveBeenCalled()
  })

  it('calls /api/auth/login endpoint in login mode', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ token: 'abc' }))

    const wrapper = mount(Login)
    await wrapper.find('input[type="password"]').setValue('validpass')
    await wrapper.find('form').trigger('submit')
    await wrapper.vm.$nextTick()

    expect(mockFetch).toHaveBeenCalledWith('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: 'validpass' }),
    })
  })

  it('calls /api/auth/setup endpoint in setup mode', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ token: 'abc' }))

    const wrapper = mount(Login, { props: { setup: true } })
    const inputs = wrapper.findAll('input[type="password"]')
    await inputs[0].setValue('validpass')
    await inputs[1].setValue('validpass')
    await wrapper.find('form').trigger('submit')
    await wrapper.vm.$nextTick()

    expect(mockFetch).toHaveBeenCalledWith('/api/auth/setup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: 'validpass' }),
    })
  })

  it('emits success on successful authentication', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ token: 'abc' }))

    const wrapper = mount(Login)
    await wrapper.find('input[type="password"]').setValue('validpass')
    await wrapper.find('form').trigger('submit')

    // Wait for the async submit to complete
    await vi.waitFor(() => {
      expect(wrapper.emitted('success')).toBeTruthy()
    })
  })

  it('shows server error message on failed login', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ error: 'Invalid password' }, 401))

    const wrapper = mount(Login)
    await wrapper.find('input[type="password"]').setValue('wrongpass')
    await wrapper.find('form').trigger('submit')

    await vi.waitFor(() => {
      expect(wrapper.text()).toContain('Invalid password')
    })
  })

  it('shows generic error when server returns no error message', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({}, 401))

    const wrapper = mount(Login)
    await wrapper.find('input[type="password"]').setValue('wrongpass')
    await wrapper.find('form').trigger('submit')

    await vi.waitFor(() => {
      expect(wrapper.text()).toContain('Authentication failed')
    })
  })
})
