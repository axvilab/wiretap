export const TraceType = {
  REQUEST: 'request',
  JOB: 'job',
  COMMAND: 'command',
  SCHEDULE: 'schedule',
} as const

export const ROOT_TRACE_TYPES = [TraceType.REQUEST, TraceType.COMMAND, TraceType.SCHEDULE] as const

export const JobStatus = {
  COMPLETED: 'completed',
  FAILED: 'failed',
  PENDING: 'pending',
  PROCESSING: 'processing',
} as const

export const EventType = {
  QUERY: 'query',
  EXCEPTION: 'exception',
  JOB_START: 'job.start',
  JOB_END: 'job.end',
  JOB_QUEUED: 'job.queued',
  MAIL_SENDING: 'mail.sending',
  MAIL_SENT: 'mail.sent',
  CACHE: 'cache',
  CACHE_WRITE: 'cache.write',
  CACHE_FORGET: 'cache.forget',
  TRANSACTION: 'transaction',
  REQUEST_START: 'request.start',
  REQUEST_END: 'request.end',
  BROADCAST: 'broadcast',
} as const
