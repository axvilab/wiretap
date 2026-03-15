import { describe, it, expect, vi, beforeEach } from 'vitest'
import { api } from '@/api'

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

describe('api.health', () => {
  it('returns health status', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ status: 'healthy' }))
    const result = await api.health()
    expect(result.status).toBe('healthy')
    expect(mockFetch).toHaveBeenCalledWith('/api/health')
  })
})

describe('api.dashboard', () => {
  it('passes period param', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ stats: {} }))
    await api.dashboard('24h')
    expect(mockFetch).toHaveBeenCalledWith('/api/dashboard?period=24h')
  })
})

describe('api.requests', () => {
  it('fetches with query params', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ traces: [], total: 0 }))
    await api.requests('type=job&page=1')
    expect(mockFetch).toHaveBeenCalledWith('/api/requests?type=job&page=1')
  })
})

describe('api.replayConfig', () => {
  it('returns enabled status', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ enabled: true }))
    const result = await api.replayConfig()
    expect(result.enabled).toBe(true)
  })
})

describe('api.replayRequest', () => {
  it('sends POST with request data', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ success: true, status: 200 }))
    const result = await api.replayRequest({ method: 'GET', uri: '/' })
    expect(result.success).toBe(true)
    expect(mockFetch).toHaveBeenCalledWith('/api/replay/request', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ method: 'GET', uri: '/' }),
    })
  })
})

describe('api.retryJob', () => {
  it('sends job UUID', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ success: true, output: 'retried' }))
    const result = await api.retryJob('abc-123')
    expect(result.success).toBe(true)
    expect(mockFetch).toHaveBeenCalledWith('/api/replay/job', expect.objectContaining({
      method: 'POST',
      body: JSON.stringify({ uuid: 'abc-123' }),
    }))
  })
})

describe('api.replayCommand', () => {
  it('sends command with input', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ success: true, exit_code: 0 }))
    const result = await api.replayCommand({ command: 'inspire', input: '' })
    expect(result.success).toBe(true)
  })
})

describe('api.dispatchJob', () => {
  it('sends serialized job data', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ success: true, job_class: 'App\\Jobs\\Test' }))
    const result = await api.dispatchJob({ job_serialized: 'base64data', job_payload: { id: 1 } })
    expect(result.success).toBe(true)
    expect(result.job_class).toBe('App\\Jobs\\Test')
  })
})

describe('api.getSettings', () => {
  it('returns settings map', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ app_url: 'http://localhost', replay_secret: 'wt_test' }))
    const result = await api.getSettings()
    expect(result.app_url).toBe('http://localhost')
  })
})

describe('api.updateSettings', () => {
  it('sends PUT with settings data', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ success: true }))
    const settings = { app_url: 'http://example.com', replay_secret: 'new_secret' }
    await api.updateSettings(settings)
    expect(mockFetch).toHaveBeenCalledWith('/api/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(settings),
    })
  })

  it('throws on non-ok response', async () => {
    mockFetch.mockReturnValueOnce(Promise.resolve({
      ok: false,
      status: 422,
      json: () => Promise.resolve({ error: 'invalid' }),
    }))
    await expect(api.updateSettings({ bad: 'data' })).rejects.toThrow('API error: 422')
  })
})

describe('api.dispatchJob', () => {
  it('sends POST to dispatch-job endpoint', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ success: true, job_class: 'App\\Jobs\\SendEmail' }))
    const payload = { job_serialized: 'serialized_data', job_payload: { user_id: 5 } }
    const result = await api.dispatchJob(payload)
    expect(result.success).toBe(true)
    expect(result.job_class).toBe('App\\Jobs\\SendEmail')
    expect(mockFetch).toHaveBeenCalledWith('/api/replay/dispatch-job', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  })

  it('throws with error message from server', async () => {
    mockFetch.mockReturnValueOnce(Promise.resolve({
      ok: false,
      status: 500,
      json: () => Promise.resolve({ error: 'job class not found' }),
    }))
    await expect(api.dispatchJob({ job_serialized: 'bad' })).rejects.toThrow('job class not found')
  })
})

describe('api error handling', () => {
  it('throws on non-ok response', async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ error: 'not found' }, 404))
    await expect(api.health()).rejects.toThrow('API error: 404')
  })
})
