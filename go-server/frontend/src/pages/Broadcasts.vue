<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Trace } from '@/api'
import { useWebSocket } from '@/hooks/useWebSocket'
import { EventType } from '@/constants'
import Pagination from '@/components/Pagination.vue'

interface BroadcastChannel {
  name: string
  type: string
}

interface BroadcastRow {
  rid: string
  ts: number
  event: string
  channels: BroadcastChannel[]
}

const broadcasts = ref<BroadcastRow[]>([])
const page = ref(1)
const perPage = ref(25)

onMounted(async () => {
  const data = await api.requests('per_page=200')
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
      if (ev.type === EventType.BROADCAST) {
        broadcasts.value.unshift({
          rid: trace.rid,
          ts: trace.ts_start + ev.ts,
          event: ev.data.event,
          channels: ev.data.channels || [],
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

const CHANNEL_STYLES: Record<string, string> = {
  public: 'bg-sky-500/10 text-sky-400 border-sky-500/20',
  private: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
  presence: 'bg-violet-500/10 text-violet-400 border-violet-500/20',
}

const paged = () => broadcasts.value.slice((page.value - 1) * perPage.value, page.value * perPage.value)
</script>

<template>
  <div class="p-6 space-y-4">
    <h1 class="text-2xl font-bold">Broadcasts ({{ broadcasts.length }})</h1>

    <div v-if="!broadcasts.length" class="text-zinc-500 py-8 text-center">No broadcasts captured yet</div>

    <div v-if="broadcasts.length" class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Event</th>
            <th class="px-4 py-2 text-left">Channels</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(b, i) in paged()"
            :key="i"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer transition-colors"
            @click="$router.push(`/broadcasts/${b.rid}`)"
          >
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs">{{ timeAgo(b.ts) }}</td>
            <td class="px-4 py-2 font-mono text-xs text-indigo-400">{{ b.event }}</td>
            <td class="px-4 py-2">
              <div class="flex gap-1 flex-wrap">
                <span
                  v-for="ch in b.channels"
                  :key="ch.name"
                  class="inline-flex items-center gap-1 px-1.5 py-0.5 text-xs rounded border"
                  :class="CHANNEL_STYLES[ch.type] || CHANNEL_STYLES.public"
                >
                  <span class="opacity-60">{{ ch.type }}</span>
                  <span class="font-mono">{{ ch.name }}</span>
                </span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="broadcasts.length" />
  </div>
</template>
