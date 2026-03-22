<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api, type Trace } from '@/api'
import { EventType } from '@/constants'
import Pagination from '@/components/Pagination.vue'
import NewItemsBanner from '@/components/NewItemsBanner.vue'

interface LogRow {
  rid: string
  eventIndex: number
  ts: number
  level: string
  message: string
  context: Record<string, any>
  traceType: string
  traceUri: string
}

const logs = ref<LogRow[]>([])
const levelFilter = ref('')
const page = ref(1)
const perPage = ref(25)

async function load() {
  logs.value = []
  const data = await api.requests('per_page=200')
  for (const t of data.traces) {
    extractLogs(t)
  }
}

function extractLogs(trace: Trace) {
  try {
    const events = JSON.parse(trace.events_json)
    events.forEach((ev: any, i: number) => {
      if (ev.type === EventType.LOG) {
        logs.value.push({
          rid: trace.rid,
          eventIndex: i,
          ts: trace.ts_start + ev.ts,
          level: ev.data.level,
          message: ev.data.message,
          context: ev.data.context || {},
          traceType: trace.type,
          traceUri: trace.uri || trace.type,
        })
      }
    })
  } catch {}
}

// Sort newest first
const sorted = computed(() => [...logs.value].sort((a, b) => b.ts - a.ts))

const filtered = computed(() => {
  if (!levelFilter.value) return sorted.value
  return sorted.value.filter(l => l.level === levelFilter.value)
})

const paged = computed(() =>
  filtered.value.slice((page.value - 1) * perPage.value, page.value * perPage.value)
)

onMounted(load)

function timeAgo(ts: number): string {
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

const LEVEL_STYLES: Record<string, string> = {
  debug: 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20',
  info: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
  notice: 'bg-sky-500/10 text-sky-400 border-sky-500/20',
  warning: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
  error: 'bg-red-500/10 text-red-400 border-red-500/20',
  critical: 'bg-red-500/10 text-red-300 border-red-500/20',
  alert: 'bg-red-500/10 text-red-300 border-red-500/20',
  emergency: 'bg-red-500/20 text-red-300 border-red-400/30',
}

const LEVELS = ['debug', 'info', 'notice', 'warning', 'error', 'critical']
</script>

<template>
  <div class="p-6 space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">Logs ({{ filtered.length }})</h1>
      <select v-model="levelFilter" class="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-sm">
        <option value="">All levels</option>
        <option v-for="l in LEVELS" :key="l" :value="l">{{ l }}</option>
      </select>
    </div>

    <NewItemsBanner :logs-only="true" @refresh="load" />

    <div v-if="!filtered.length" class="text-zinc-500 py-8 text-center">No logs captured yet</div>

    <div v-if="filtered.length" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Level</th>
            <th class="px-4 py-2 text-left">Message</th>
            <th class="px-4 py-2 text-left">Trace</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(l, i) in paged"
            :key="i"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer transition-colors"
            @click="$router.push(`/logs/${l.rid}/${l.eventIndex}`)"
          >
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs whitespace-nowrap">{{ timeAgo(l.ts) }}</td>
            <td class="px-4 py-2">
              <span
                class="inline-flex px-1.5 py-0.5 text-xs rounded border"
                :class="LEVEL_STYLES[l.level] || LEVEL_STYLES.info"
              >{{ l.level }}</span>
            </td>
            <td class="px-4 py-2 text-xs text-zinc-300 truncate max-w-md">{{ l.message }}</td>
            <td class="px-4 py-2 text-xs text-zinc-500 font-mono truncate max-w-[200px]">{{ l.traceUri }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="filtered.length" />
  </div>
</template>
