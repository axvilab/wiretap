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
  COMMAND_START: 'command.start',
  COMMAND_END: 'command.end',
  SCHEDULE_START: 'schedule.start',
  SCHEDULE_END: 'schedule.end',
  SCHEDULE_FAILED: 'schedule.failed',
} as const

export const EVENT_TYPE_COLORS: Record<string, string> = {
  [EventType.QUERY]: 'text-sky-400',
  [EventType.REQUEST_START]: 'text-zinc-400',
  [EventType.REQUEST_END]: 'text-zinc-400',
  [EventType.EXCEPTION]: 'text-red-400',
  [EventType.JOB_QUEUED]: 'text-purple-400',
  [EventType.JOB_START]: 'text-purple-400',
  [EventType.JOB_END]: 'text-purple-400',
  [EventType.CACHE]: 'text-amber-400',
  [EventType.CACHE_WRITE]: 'text-amber-400',
  [EventType.CACHE_FORGET]: 'text-amber-400',
  [EventType.MAIL_SENDING]: 'text-emerald-400',
  [EventType.MAIL_SENT]: 'text-emerald-400',
  [EventType.BROADCAST]: 'text-indigo-400',
  [EventType.TRANSACTION]: 'text-zinc-500',
  [EventType.COMMAND_START]: 'text-orange-400',
  [EventType.COMMAND_END]: 'text-orange-400',
  [EventType.SCHEDULE_START]: 'text-violet-400',
  [EventType.SCHEDULE_END]: 'text-violet-400',
  [EventType.SCHEDULE_FAILED]: 'text-red-400',
}
