const BASE = '/api'

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`)
  if (!res.ok) throw new Error(`API error: ${res.status}`)
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
  attachments: { name: string; content_type: string; size: number }[]
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
}

export const api = {
  dashboard: (period = '1h') => get<DashboardData>(`/dashboard?period=${period}`),
  requests: (params = '') => get<ListResult>(`/requests${params ? '?' + params : ''}`),
  trace: (rid: string) => get<TraceNode>(`/trace/${rid}`),
  mail: (params = '') => get<MailListResult>(`/mail${params ? '?' + params : ''}`),
  mailDetail: (id: string) => get<Email>(`/mail/${id}`),
  mailByTrace: (rid: string) => get<Email>(`/mail/by-trace/${rid}`),
  health: () => get<{ status: string }>('/health'),
}
