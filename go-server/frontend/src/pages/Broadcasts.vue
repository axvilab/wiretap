<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type Trace } from '@/api'
import { useWebSocket } from '@/hooks/useWebSocket'

interface BroadcastEvent {
  rid: string
  ts: number
  event: string
  channels: string[]
  payload: any
}

const broadcasts = ref<BroadcastEvent[]>([])

onMounted(async () => {
  // Load recent traces and extract broadcast events
  const data = await api.requests('per_page=100')
  for (const t of data.traces) {
    extractBroadcasts(t)
  }
})

useWebSocket((event) => {
  if (event.type === 'trace.new' && event.data) {
    extractBroadcasts(event.data)
  }
})

function extractBroadcasts(trace: Trace) {
  try {
    const events = JSON.parse(trace.events_json)
    for (const ev of events) {
      if (ev.type === 'broadcast') {
        broadcasts.value.unshift({
          rid: trace.rid,
          ts: trace.ts_start + ev.ts,
          event: ev.data.event,
          channels: ev.data.channels || [],
          payload: ev.data.payload,
        })
      }
    }
  } catch {}
}

function timeAgo(ts: number): string {
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}
</script>

<template>
  <div class="p-6 space-y-4">
    <h1 class="text-2xl font-bold">Broadcasts</h1>

    <div v-if="!broadcasts.length" class="text-zinc-500 py-8 text-center">No broadcasts captured yet</div>

    <div class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <table v-if="broadcasts.length" class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Event</th>
            <th class="px-4 py-2 text-left">Channels</th>
            <th class="px-4 py-2 text-left">Payload</th>
            <th class="px-4 py-2 text-left">Trace</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(b, i) in broadcasts"
            :key="i"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30"
          >
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs">{{ timeAgo(b.ts) }}</td>
            <td class="px-4 py-2 font-mono text-xs text-indigo-400">{{ b.event }}</td>
            <td class="px-4 py-2 text-xs text-zinc-400">{{ b.channels.join(', ') }}</td>
            <td class="px-4 py-2 font-mono text-xs text-zinc-400 max-w-xs truncate">{{ JSON.stringify(b.payload) }}</td>
            <td class="px-4 py-2">
              <router-link :to="`/trace/${b.rid}`" class="text-emerald-400 text-xs hover:underline">View</router-link>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
