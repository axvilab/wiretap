<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import type { TraceNode } from '@/api'
import { EventType, TraceType, JobStatus, EVENT_TYPE_COLORS } from '@/constants'

const props = defineProps<{ node: TraceNode; depth?: number }>()
const expanded = ref((props.depth || 0) === 0)
const expandedEvents = ref<Set<number>>(new Set())
const router = useRouter()

function toggleEvent(index: number) {
  if (expandedEvents.value.has(index)) {
    expandedEvents.value.delete(index)
  } else {
    expandedEvents.value.add(index)
  }
}

function shortPath(file: string): string {
  const parts = file.split('/')
  return parts.slice(-2).join('/')
}

function parseEvents(json: string) {
  try { return JSON.parse(json) } catch { return [] }
}

function formatTs(ts: number): string {
  return (ts * 1000).toFixed(1) + 'ms'
}

function eventColor(type: string): string {
  return EVENT_TYPE_COLORS[type] || 'text-zinc-400'
}

function parseMeta(json: string) {
  try { return JSON.parse(json) } catch { return {} }
}

const STATUS_STYLES: Record<string, string> = {
  [JobStatus.COMPLETED]: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
  [JobStatus.FAILED]: 'bg-red-500/10 text-red-400 border-red-500/20',
  [JobStatus.PENDING]: 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20',
  [JobStatus.PROCESSING]: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
}

function statusBadge(status: string): string {
  return STATUS_STYLES[status] || STATUS_STYLES.processing
}

async function goToMail(traceRID: string) {
  try {
    const res = await fetch(`/api/mail/by-trace/${traceRID}`)
    if (res.ok) {
      const email = await res.json()
      router.push(`/mail/${email.id}`)
    }
  } catch {}
}
</script>

<template>
  <div>
    <!-- Node header -->
    <div
      class="flex items-center gap-2 px-4 py-2 border-b border-zinc-800 transition-colors"
      :class="[
        (depth || 0) > 0 ? 'bg-zinc-800/30 hover:bg-zinc-800/50 cursor-pointer' : 'bg-zinc-800/60',
      ]"
      :style="{ paddingLeft: `${(depth || 0) * 24 + 16}px` }"
      @click="(depth || 0) > 0 && (expanded = !expanded)"
    >
      <span v-if="(depth || 0) > 0" class="text-zinc-400 text-base w-5">{{ expanded ? '▾' : '▸' }}</span>
      <span v-if="node.trace.type === TraceType.REQUEST" class="text-sky-400 text-xs font-bold">{{ node.trace.method }}</span>
      <span v-else class="text-purple-400 text-xs font-bold">{{ node.trace.type }}</span>
      <span class="font-mono text-sm text-zinc-200">
        {{ node.trace.uri || parseMeta(node.trace.meta_json).job_class || '' }}
      </span>
      <span v-if="node.trace.status" class="text-xs px-1.5 py-0.5 rounded border"
        :class="node.trace.status >= 500 ? 'bg-red-500/10 text-red-400 border-red-500/20' :
                node.trace.status >= 400 ? 'bg-amber-500/10 text-amber-400 border-amber-500/20' :
                'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'"
      >{{ node.trace.status }}</span>
      <span class="text-zinc-500 text-xs ml-auto">{{ node.trace.duration_ms.toFixed(1) }}ms</span>
      <span class="text-zinc-600 text-xs">{{ node.trace.query_count }} queries</span>
    </div>

    <!-- Events -->
    <template v-if="expanded" v-for="(ev, i) in parseEvents(node.trace.events_json)" :key="i">
      <div
        class="flex items-start gap-3 px-4 py-1.5 text-sm hover:bg-zinc-800/20 border-b border-zinc-800/30"
        :style="{ paddingLeft: `${(depth || 0) * 24 + 28}px` }"
      >
        <span class="font-mono text-xs text-zinc-500 w-16 shrink-0 text-right">{{ formatTs(ev.ts) }}</span>
        <span class="font-mono text-xs font-bold w-28 shrink-0" :class="eventColor(ev.type)">{{ ev.type }}</span>
        <div class="flex-1 font-mono text-xs text-zinc-300 min-w-0">
          <!-- Query -->
          <template v-if="ev.type === EventType.QUERY">
            <div class="flex items-start gap-2">
              <span class="text-sky-300 truncate" :class="expandedEvents.has(i) ? '!whitespace-pre-wrap !break-all' : ''">{{ expandedEvents.has(i) ? ev.data.sql : ev.data.sql.substring(0, 80) }}{{ !expandedEvents.has(i) && ev.data.sql.length > 80 ? '...' : '' }}</span>
              <span class="text-zinc-500 shrink-0">{{ ev.data.time_ms?.toFixed(2) }}ms</span>
              <button v-if="ev.data.sql.length > 80 || (ev.data.bindings && ev.data.bindings.length)" @click.stop="toggleEvent(i)" class="px-1.5 py-0.5 text-xs text-zinc-500 hover:text-zinc-300 shrink-0">{{ expandedEvents.has(i) ? '▾' : '▸' }}</button>
            </div>
            <div v-if="expandedEvents.has(i) && ev.data.bindings && ev.data.bindings.length" class="mt-1 bg-zinc-950 rounded p-2 text-xs">
              <span class="text-zinc-500">bindings:</span>
              <span class="text-amber-300 ml-1">{{ JSON.stringify(ev.data.bindings) }}</span>
            </div>
          </template>
          <!-- Exception -->
          <template v-else-if="ev.type === EventType.EXCEPTION">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="text-red-300">{{ ev.data.class }}: {{ ev.data.message }}</span>
              <button @click.stop="router.push('/exceptions')" class="px-2 py-0.5 text-xs text-zinc-400 bg-zinc-800 rounded hover:text-zinc-200 hover:bg-zinc-700 transition-colors">Open</button>
            </div>
          </template>
          <!-- Mail sent -->
          <template v-else-if="ev.type === EventType.MAIL_SENT">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="text-emerald-300">{{ ev.data.subject }}</span>
              <span class="text-zinc-500">→ {{ ev.data.to?.join(', ') }}</span>
              <span v-if="ev.data.attachments?.length" class="text-zinc-600">{{ ev.data.attachments.length }} attachment(s)</span>
              <button @click.stop="goToMail(node.trace.rid)" class="px-2 py-0.5 text-xs text-zinc-400 bg-zinc-800 rounded hover:text-zinc-200 hover:bg-zinc-700 transition-colors">Open</button>
            </div>
          </template>
          <!-- Mail sending -->
          <template v-else-if="ev.type === EventType.MAIL_SENDING">
            <span class="text-emerald-400">{{ ev.data.subject }}</span>
            <span class="text-zinc-500 ml-2">→ {{ ev.data.to?.join(', ') }}</span>
          </template>
          <!-- Job queued -->
          <template v-else-if="ev.type === EventType.JOB_QUEUED">
            <span class="text-purple-300">{{ ev.data.job_class }}</span>
            <span class="text-zinc-500 ml-2">queue: {{ ev.data.queue }}</span>
          </template>
          <!-- Job start/end -->
          <template v-else-if="ev.type === EventType.JOB_START || ev.type === EventType.JOB_END">
            <span class="text-purple-300">{{ ev.data.job_class }}</span>
            <span v-if="ev.data.status" class="ml-2 px-1.5 py-0.5 rounded text-xs border" :class="statusBadge(ev.data.status)">{{ ev.data.status }}</span>
          </template>
          <!-- Cache -->
          <template v-else-if="ev.type === EventType.CACHE || ev.type === EventType.CACHE_WRITE || ev.type === EventType.CACHE_FORGET">
            <span class="text-amber-300">{{ ev.data.key }}</span>
            <span v-if="ev.data.hit !== undefined" class="text-zinc-500 ml-2">{{ ev.data.hit ? 'HIT' : 'MISS' }}</span>
          </template>
          <!-- Broadcast -->
          <template v-else-if="ev.type === EventType.BROADCAST">
            <div class="flex items-center gap-1 flex-wrap">
              <span class="text-indigo-300">{{ ev.data.event }}</span>
              <span v-for="ch in ev.data.channels" :key="ch.name" class="px-1 py-0.5 rounded border text-xs"
                :class="ch.type === 'presence' ? 'bg-violet-500/10 text-violet-400 border-violet-500/20' : ch.type === 'private' ? 'bg-amber-500/10 text-amber-400 border-amber-500/20' : 'bg-sky-500/10 text-sky-400 border-sky-500/20'"
              >{{ ch.type }} {{ ch.name }}</span>
              <button @click.stop="router.push(`/broadcasts/${node.trace.rid}`)" class="px-2 py-0.5 text-xs text-zinc-400 bg-zinc-800 rounded hover:text-zinc-200 hover:bg-zinc-700 transition-colors">Open</button>
            </div>
          </template>
          <!-- Transaction -->
          <template v-else-if="ev.type === EventType.TRANSACTION">
            <span class="text-zinc-400">{{ ev.data.action }}</span>
            <span class="text-zinc-600 ml-2">{{ ev.data.connection }}</span>
          </template>
          <!-- Command -->
          <template v-else-if="ev.type === EventType.COMMAND_START">
            <span class="text-orange-300">{{ ev.data.command }}</span>
          </template>
          <template v-else-if="ev.type === EventType.COMMAND_END">
            <span class="text-orange-300">{{ ev.data.command }}</span>
            <span class="ml-2 px-1.5 py-0.5 rounded text-xs border" :class="ev.data.exit_code === 0 ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-red-500/10 text-red-400 border-red-500/20'">exit {{ ev.data.exit_code }}</span>
          </template>
          <!-- Schedule -->
          <template v-else-if="ev.type === EventType.SCHEDULE_START">
            <span class="text-violet-300">{{ ev.data.command }}</span>
            <span class="text-zinc-500 ml-2">{{ ev.data.expression }}</span>
          </template>
          <template v-else-if="ev.type === EventType.SCHEDULE_END">
            <span class="text-violet-300">{{ ev.data.command }}</span>
            <span class="ml-2 px-1.5 py-0.5 rounded text-xs border bg-emerald-500/10 text-emerald-400 border-emerald-500/20">done</span>
          </template>
          <template v-else-if="ev.type === EventType.SCHEDULE_FAILED">
            <span class="text-violet-300">{{ ev.data.expression }}</span>
            <span class="text-red-400 ml-2">{{ ev.data.error }}</span>
          </template>
          <!-- Notification -->
          <template v-else-if="ev.type === EventType.NOTIFICATION_SENDING || ev.type === EventType.NOTIFICATION_SENT">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="text-teal-300">{{ ev.data.class }}</span>
              <span class="px-1.5 py-0.5 rounded border text-xs" :class="ev.data.channel === 'mail' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20'">{{ ev.data.channel }}</span>
              <span class="text-zinc-500">→ {{ ev.data.notifiable }}</span>
              <button @click.stop="router.push(`/notifications/${node.trace.rid}/${i}`)" class="px-2 py-0.5 text-xs text-zinc-400 bg-zinc-800 rounded hover:text-zinc-200 hover:bg-zinc-700 transition-colors">Open</button>
            </div>
          </template>
          <!-- Log -->
          <template v-else-if="ev.type === EventType.LOG">
            <div class="flex items-center gap-2">
              <span class="px-1.5 py-0.5 rounded border text-xs shrink-0" :class="{ 'bg-red-500/10 text-red-400 border-red-500/20': ['error','critical','alert','emergency'].includes(ev.data.level), 'bg-amber-500/10 text-amber-400 border-amber-500/20': ev.data.level === 'warning', 'bg-blue-500/10 text-blue-400 border-blue-500/20': ['info','notice'].includes(ev.data.level), 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20': ev.data.level === 'debug' }">{{ ev.data.level }}</span>
              <span class="text-zinc-300 truncate">{{ ev.data.message.substring(0, 80) }}{{ ev.data.message.length > 80 ? '...' : '' }}</span>
              <button @click.stop="router.push(`/logs/${node.trace.rid}/${i}`)" class="px-2 py-0.5 text-xs text-zinc-400 bg-zinc-800 rounded hover:text-zinc-200 hover:bg-zinc-700 transition-colors shrink-0">Open</button>
            </div>
          </template>
          <!-- Dump -->
          <template v-else-if="ev.type === EventType.DUMP">
            <div class="flex items-start gap-2">
              <span class="text-zinc-500 shrink-0">{{ shortPath(ev.data.file) }}:{{ ev.data.line }}</span>
              <button @click.stop="toggleEvent(i)" class="px-1.5 py-0.5 text-xs text-zinc-500 hover:text-zinc-300 shrink-0">{{ expandedEvents.has(i) ? '▾' : '▸' }}</button>
            </div>
            <pre v-if="!expandedEvents.has(i)" class="mt-0.5 text-pink-300/70 line-clamp-1">{{ ev.data.value.substring(0, 80) }}</pre>
            <pre v-else class="mt-1 text-pink-300 bg-zinc-950 rounded p-2 whitespace-pre-wrap break-words max-h-[300px] overflow-auto">{{ ev.data.value }}</pre>
          </template>
          <!-- Validation -->
          <template v-else-if="ev.type === EventType.VALIDATION">
            <div class="flex items-center gap-2">
              <span class="text-red-400 font-bold shrink-0">Validation failed</span>
              <span class="text-amber-300 shrink-0">{{ Object.keys(ev.data.errors).join(', ') }}</span>
              <button @click.stop="toggleEvent(i)" class="px-1.5 py-0.5 text-xs text-zinc-500 hover:text-zinc-300 shrink-0">{{ expandedEvents.has(i) ? '▾' : '▸' }}</button>
            </div>
            <div v-if="expandedEvents.has(i)" class="mt-1 bg-zinc-950 rounded p-2 text-xs space-y-1">
              <div v-for="(msgs, field) in ev.data.errors" :key="field">
                <span class="text-amber-300">{{ field }}:</span>
                <span class="text-zinc-400 ml-1">{{ Array.isArray(msgs) ? msgs.join(', ') : msgs }}</span>
              </div>
            </div>
          </template>
          <!-- Auth -->
          <template v-else-if="ev.type === EventType.AUTH_LOGIN">
            <span class="px-1.5 py-0.5 rounded border text-xs bg-emerald-500/10 text-emerald-400 border-emerald-500/20">login</span>
            <span class="ml-2 text-zinc-300">user #{{ ev.data.user_id }}</span>
            <span class="ml-2 text-zinc-500">guard: {{ ev.data.guard }}</span>
          </template>
          <template v-else-if="ev.type === EventType.AUTH_LOGOUT">
            <span class="px-1.5 py-0.5 rounded border text-xs bg-zinc-500/10 text-zinc-400 border-zinc-500/20">logout</span>
            <span v-if="ev.data.user_id" class="ml-2 text-zinc-300">user #{{ ev.data.user_id }}</span>
            <span class="ml-2 text-zinc-500">guard: {{ ev.data.guard }}</span>
          </template>
          <template v-else-if="ev.type === EventType.AUTH_FAILED">
            <span class="px-1.5 py-0.5 rounded border text-xs bg-red-500/10 text-red-400 border-red-500/20">failed</span>
            <span class="ml-2 text-zinc-500">guard: {{ ev.data.guard }}</span>
            <span class="ml-2 text-zinc-500">fields: {{ ev.data.credentials?.join(', ') }}</span>
          </template>
          <!-- Gate -->
          <template v-else-if="ev.type === EventType.GATE">
            <span class="px-1.5 py-0.5 rounded border text-xs"
              :class="ev.data.result ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-red-500/10 text-red-400 border-red-500/20'"
            >{{ ev.data.result ? 'allowed' : 'denied' }}</span>
            <span class="ml-2 text-zinc-300">{{ ev.data.ability }}</span>
            <span v-if="ev.data.user_id" class="ml-2 text-zinc-500">user #{{ ev.data.user_id }}</span>
          </template>
          <!-- Http Client -->
          <template v-else-if="ev.type === EventType.HTTP_OUT_START">
            <div class="flex items-center gap-2">
              <span class="text-sky-300 shrink-0">{{ ev.data.method }}</span>
              <span class="text-zinc-300 truncate">{{ ev.data.url?.substring(0, 80) }}{{ ev.data.url?.length > 80 ? '...' : '' }}</span>
              <button v-if="ev.data.headers || ev.data.body" @click.stop="toggleEvent(i)" class="px-1.5 py-0.5 text-xs text-zinc-500 hover:text-zinc-300 shrink-0">{{ expandedEvents.has(i) ? '▾' : '▸' }}</button>
            </div>
            <div v-if="expandedEvents.has(i)" class="mt-1 bg-zinc-950 rounded p-2 text-xs space-y-1 max-h-[200px] overflow-auto">
              <div v-if="ev.data.url?.length > 80"><span class="text-zinc-500">url:</span> <span class="text-zinc-300 break-all">{{ ev.data.url }}</span></div>
              <div v-if="ev.data.headers"><span class="text-zinc-500">headers:</span> <pre class="text-zinc-400 whitespace-pre-wrap">{{ JSON.stringify(ev.data.headers, null, 2) }}</pre></div>
              <div v-if="ev.data.body"><span class="text-zinc-500">body:</span> <pre class="text-amber-300 whitespace-pre-wrap">{{ ev.data.body }}</pre></div>
            </div>
          </template>
          <template v-else-if="ev.type === EventType.HTTP_OUT_END">
            <div class="flex items-center gap-2">
              <span class="px-1.5 py-0.5 rounded border text-xs"
                :class="ev.data.status < 400 ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-red-500/10 text-red-400 border-red-500/20'"
              >{{ ev.data.status }}</span>
              <span v-if="ev.data.response_size" class="text-zinc-500">{{ (ev.data.response_size / 1024).toFixed(1) }}KB</span>
              <span v-if="ev.data.response_body_skipped" class="text-zinc-600">{{ ev.data.response_body_skipped }}</span>
              <button v-if="ev.data.response_body || ev.data.response_headers" @click.stop="toggleEvent(i)" class="px-1.5 py-0.5 text-xs text-zinc-500 hover:text-zinc-300 shrink-0">{{ expandedEvents.has(i) ? '▾' : '▸' }}</button>
            </div>
            <div v-if="expandedEvents.has(i)" class="mt-1 bg-zinc-950 rounded p-2 text-xs space-y-1 max-h-[300px] overflow-auto">
              <div v-if="ev.data.response_headers"><span class="text-zinc-500">headers:</span> <pre class="text-zinc-400 whitespace-pre-wrap">{{ JSON.stringify(ev.data.response_headers, null, 2) }}</pre></div>
              <div v-if="ev.data.response_body"><span class="text-zinc-500">body:</span> <pre class="text-emerald-300 whitespace-pre-wrap break-words">{{ ev.data.response_body.substring(0, 2000) }}{{ ev.data.response_body.length > 2000 ? '\n... (truncated in view)' : '' }}</pre></div>
              <div v-if="ev.data.response_truncated" class="text-amber-400">Response truncated (original: {{ (ev.data.response_size / 1024).toFixed(1) }}KB)</div>
            </div>
          </template>
          <!-- Model -->
          <template v-else-if="ev.type === EventType.MODEL">
            <span class="px-1.5 py-0.5 rounded border text-xs shrink-0"
              :class="{
                'bg-emerald-500/10 text-emerald-400 border-emerald-500/20': ev.data.action === 'created',
                'bg-blue-500/10 text-blue-400 border-blue-500/20': ev.data.action === 'updated',
                'bg-red-500/10 text-red-400 border-red-500/20': ev.data.action === 'deleted',
              }"
            >{{ ev.data.action }}</span>
            <span class="ml-2 text-cyan-300 truncate">{{ ev.data.model.length > 60 ? '...' + ev.data.model.slice(-50) : ev.data.model }}</span>
            <span v-if="ev.data.key" class="ml-1 text-zinc-500 shrink-0">#{{ ev.data.key }}</span>
          </template>
          <!-- View -->
          <template v-else-if="ev.type === EventType.VIEW">
            <span class="text-lime-300">{{ ev.data.name }}</span>
            <span v-if="ev.data.path" class="ml-2 text-zinc-600 truncate">{{ ev.data.path.length > 60 ? '...' + ev.data.path.slice(-50) : ev.data.path }}</span>
          </template>
          <!-- Request start -->
          <template v-else-if="ev.type === EventType.REQUEST_START">
            <div class="flex items-center gap-2">
              <span class="text-zinc-300">{{ ev.data.method }} {{ ev.data.uri }}</span>
              <span class="text-zinc-600">{{ ev.data.ip }}</span>
              <button v-if="ev.data.headers || ev.data.body" @click.stop="toggleEvent(i)" class="px-1.5 py-0.5 text-xs text-zinc-500 hover:text-zinc-300 shrink-0">{{ expandedEvents.has(i) ? '▾' : '▸' }}</button>
            </div>
            <div v-if="expandedEvents.has(i)" class="mt-1 bg-zinc-950 rounded p-2 text-xs space-y-1 max-h-[200px] overflow-auto">
              <div v-if="ev.data.headers"><span class="text-zinc-500">headers:</span> <span class="text-zinc-400">{{ JSON.stringify(ev.data.headers, null, 2) }}</span></div>
              <div v-if="ev.data.body"><span class="text-zinc-500">body:</span> <span class="text-amber-300">{{ JSON.stringify(ev.data.body, null, 2) }}</span></div>
            </div>
          </template>
          <!-- Request end -->
          <template v-else-if="ev.type === EventType.REQUEST_END">
            <span class="px-1.5 py-0.5 rounded border text-xs" :class="ev.data.status < 400 ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : ev.data.status < 500 ? 'bg-amber-500/10 text-amber-400 border-amber-500/20' : 'bg-red-500/10 text-red-400 border-red-500/20'">{{ ev.data.status }}</span>
            <span class="ml-2 text-zinc-400">{{ ev.data.duration_ms?.toFixed(1) }}ms</span>
            <span class="ml-2 text-zinc-600">{{ ev.data.memory_mb }}MB</span>
          </template>
          <!-- Default -->
          <template v-else>
            <div class="flex items-center gap-2">
              <span class="text-zinc-400 truncate">{{ JSON.stringify(ev.data).substring(0, 80) }}{{ JSON.stringify(ev.data).length > 80 ? '...' : '' }}</span>
              <button v-if="JSON.stringify(ev.data).length > 80" @click.stop="toggleEvent(i)" class="px-1.5 py-0.5 text-xs text-zinc-500 hover:text-zinc-300 shrink-0">{{ expandedEvents.has(i) ? '▾' : '▸' }}</button>
            </div>
            <pre v-if="expandedEvents.has(i)" class="mt-1 bg-zinc-950 rounded p-2 text-xs text-zinc-400 whitespace-pre-wrap break-words max-h-[200px] overflow-auto">{{ JSON.stringify(ev.data, null, 2) }}</pre>
          </template>
        </div>
      </div>
    </template>

    <!-- Recursive children (visible when parent expanded) -->
    <TraceTree
      v-if="expanded"
      v-for="child in node.children"
      :key="child.trace.rid"
      :node="child"
      :depth="(depth || 0) + 1"
    />
  </div>
</template>
