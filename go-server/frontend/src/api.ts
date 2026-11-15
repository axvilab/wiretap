const BASE = '/api'

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`)
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json()
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: `HTTP ${res.status}` }))
    throw new Error(data.error || `API error: ${res.status}`)
  }
  return res.json()
}

export interface Trace {
  rid: string
  parent_rid: string
  type: string
  method: string
  uri: string
  status: number
  duration_ms: number
  ts_start: number
  ts_end: number
  query_count: number
  event_count: number
  job_count: number
  has_exception: boolean
  has_n_plus_one: boolean
  has_slow_query: boolean
  has_duplicate_query: boolean
  meta_json: string
  events_json: string
  warnings_json: string
}

export interface TraceNode {
  trace: Trace
  children: TraceNode[]
}

export interface ListResult {
  traces: Trace[]
  total: number
  page: number
  per_page: number
}

export interface Email {
  id: string
  from: string
  to: string[]
  cc: string[]
  subject: string
  html: string
  text: string
  attachments: { index: number; name: string; content_type: string; size: number }[]
  raw_size: number
  trace_rid: string
  created_at: number
}

export interface MailListResult {
  emails: Email[]
  total: number
  page: number
  per_page: number
}

export interface DashboardData {
  period: string
  stats: {
    total_requests: number
    total_jobs: number
    total_emails: number
    total_exceptions: number
    avg_duration: number
    median_duration: number
    p95_duration: number
    requests_per_min: number
  }
  timeline: { timestamp: string; count: number; avg_ms: number; p95_ms: number; error_count: number }[]
  status_breakdown: { group: string; count: number }[]
  slow_endpoints: { method: string; uri: string; count: number; avg_ms: number; max_ms: number; p95_ms: number }[]
  slow_queries: { sql: string; time_ms: number; trace_rid: string; trace_type: string; trace_uri: string }[]
}

async function put<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json()
}

async function del<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json()
}

// --- Monitor types ---

export interface MonitorSite {
  id: number
  url: string
  name: string
  check_ssl: boolean
  health_url: string
  auth_type: string
  auth_json: string
  interval_s: number
  is_primary: boolean
}

export interface MonitorCheck {
  id: number
  site_id: number
  status: number
  response_ms: number
  error: string
  checked_at: number
}

export interface MonitorSSL {
  site_id: number
  issuer: string
  subject: string
  expires_at: number
  days_remaining: number
  is_valid: boolean
  error: string
  checked_at: number
}

export interface MonitorHealth {
  site_id: number
  status: string
  data_json: string
  checked_at: number
}

export interface MonitorDomain {
  site_id: number
  domain: string
  registrar: string
  expires_at: number
  days_remaining: number
  error: string
  checked_at: number
}

export interface MonitorStatus {
  site: MonitorSite
  ssl: MonitorSSL | null
  health: MonitorHealth | null
  domain: MonitorDomain | null
  last_check: MonitorCheck | null
  uptime_pct: number
}

export interface NotificationChannel {
  id: number
  type: string
  name: string
  config: string
  enabled: boolean
}

export interface AlertRule {
  id: number
  type: string
  channel_id: number
  config: string
  enabled: boolean
  cooldown_m: number
}

export interface AlertLog {
  id: number
  rule_id: number
  message: string
  sent_at: number
}

export interface ScheduleHeartbeat {
  id: number
  command: string
  mode: string
  max_gap_m: number
  at_time: string
  grace_m: number
  last_seen_at: number
  alerted: boolean
}

export const api = {
  dashboard: (period = '1h') => get<DashboardData>(`/dashboard?period=${period}`),
  requests: (params = '') => get<ListResult>(`/requests${params ? '?' + params : ''}`),
  trace: (rid: string) => get<TraceNode>(`/trace/${rid}`),
  mail: (params = '') => get<MailListResult>(`/mail${params ? '?' + params : ''}`),
  mailDetail: (id: string) => get<Email>(`/mail/${id}`),
  mailByTrace: (rid: string) => get<Email>(`/mail/by-trace/${rid}`),
  health: () => get<{ status: string }>('/health'),

  getSettings: () => get<Record<string, string>>('/settings'),
  updateSettings: (data: Record<string, string>) =>
    fetch(`${BASE}/settings`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    }).then(r => { if (!r.ok) throw new Error(`API error: ${r.status}`); return r.json() }),

  replayConfig: () => get<{ enabled: boolean }>('/replay/config'),
  replayRequest: (data: { method: string; uri: string; headers?: Record<string, string>; body?: string }) =>
    post<{ success: boolean; status?: number; error?: string }>('/replay/request', data),
  retryJob: (uuid: string) =>
    post<{ success: boolean; output?: string; error?: string }>('/replay/job', { uuid }),
  dispatchJob: (data: { job_serialized: string; job_payload?: Record<string, unknown> }) =>
    post<{ success: boolean; job_class?: string; error?: string }>('/replay/dispatch-job', data),
  replayBroadcast: (data: { event: string; channels: { name: string; type: string }[]; payload: Record<string, unknown> }) =>
    post<{ success: boolean; error?: string }>('/replay/broadcast', data),
  replayCommand: (data: { command: string; input?: string }) =>
    post<{ success: boolean; output?: string; exit_code?: number; error?: string }>('/replay/command', data),

  // --- Monitor ---
  monitorStatus: () => get<MonitorStatus[]>('/monitor/status'),
  monitorSites: () => get<MonitorSite[]>('/monitor/sites'),
  monitorCreateSite: (site: Partial<MonitorSite>) => post<MonitorSite>('/monitor/sites', site),
  monitorUpdateSite: (id: number, site: Partial<MonitorSite>) => put<any>(`/monitor/sites/${id}`, site),
  monitorDeleteSite: (id: number) => del<any>(`/monitor/sites/${id}`),
  monitorChecks: (siteId: number, limit = 100) => get<MonitorCheck[]>(`/monitor/checks/${siteId}?limit=${limit}`),
  monitorSSL: (siteId: number) => get<MonitorSSL | null>(`/monitor/ssl/${siteId}`),
  monitorHealth: (siteId: number) => get<MonitorHealth | null>(`/monitor/health/${siteId}`),
  monitorDomain: (siteId: number) => get<MonitorDomain | null>(`/monitor/domain/${siteId}`),

  // --- Schedule Heartbeats ---
  knownScheduleCommands: () => get<string[]>('/monitor/schedule-commands'),
  scheduleHeartbeats: () => get<ScheduleHeartbeat[]>('/monitor/schedules'),
  createScheduleHeartbeat: (b: Partial<ScheduleHeartbeat>) => post<ScheduleHeartbeat>('/monitor/schedules', b),
  updateScheduleHeartbeat: (id: number, b: Partial<ScheduleHeartbeat>) => put<any>(`/monitor/schedules/${id}`, b),
  deleteScheduleHeartbeat: (id: number) => del<any>(`/monitor/schedules/${id}`),

  // --- Notifications ---
  notificationChannels: () => get<NotificationChannel[]>('/notifications/channels'),
  createNotificationChannel: (ch: Partial<NotificationChannel>) => post<NotificationChannel>('/notifications/channels', ch),
  updateNotificationChannel: (id: number, ch: Partial<NotificationChannel>) => put<any>(`/notifications/channels/${id}`, ch),
  deleteNotificationChannel: (id: number) => del<any>(`/notifications/channels/${id}`),
  testNotificationChannel: (id: number) => post<any>(`/notifications/channels/${id}/test`, {}),

  alertRules: () => get<AlertRule[]>('/notifications/rules'),
  createAlertRule: (rule: Partial<AlertRule>) => post<AlertRule>('/notifications/rules', rule),
  updateAlertRule: (id: number, rule: Partial<AlertRule>) => put<any>(`/notifications/rules/${id}`, rule),
  deleteAlertRule: (id: number) => del<any>(`/notifications/rules/${id}`),

  alertLogs: (limit = 50) => get<AlertLog[]>(`/notifications/logs?limit=${limit}`),
}
