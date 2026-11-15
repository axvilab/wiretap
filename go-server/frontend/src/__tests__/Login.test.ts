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
  it('shows "Sign in to continue"', () => {
    const wrapper = mount(Login)
    expect(wrapper.text()).toContain('Sign in to continue')
  })

  it('shows button text "Sign In"', () => {
    const wrapper = mount(Login)
    expect(wrapper.find('button[type="submit"]').text()).toBe('Sign In')
  })

  it('shows an error and skips the request when the password is empty', async () => {
    const wrapper = mount(Login)

    await wrapper.find('form').trigger('submit')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Password is required')
    expect(mockFetch).not.toHaveBeenCalled()
  })

  it('calls the /api/auth/login endpoint', async () => {
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
