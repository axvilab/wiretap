<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api, type Trace } from '@/api'
import Pagination from '@/components/Pagination.vue'
import NewItemsBanner from '@/components/NewItemsBanner.vue'

interface NotificationRow {
  rid: string
  eventIndex: number
  ts: number
  channel: string
  class: string
  notifiable: string
  status: 'sending' | 'sent'
  traceUri: string
}

const notifications = ref<NotificationRow[]>([])
const page = ref(1)
const perPage = ref(25)

async function load() {
  notifications.value = []
  const data = await api.requests('per_page=200')
  for (const t of data.traces) {
    extractNotifications(t)
  }
}

function extractNotifications(trace: Trace) {
  try {
    const events = JSON.parse(trace.events_json)
    events.forEach((ev: any, i: number) => {
      if (ev.type === 'notification.sent' || ev.type === 'notification.sending') {
        notifications.value.push({
          rid: trace.rid,
          eventIndex: i,
          ts: trace.ts_start + ev.ts,
          channel: ev.data.channel,
          class: ev.data.class,
          notifiable: ev.data.notifiable,
          status: ev.type === 'notification.sent' ? 'sent' : 'sending',
          traceUri: trace.uri || trace.type,
        })
      }
    })
  } catch {}
}

const sorted = computed(() => [...notifications.value].sort((a, b) => b.ts - a.ts))
// Deduplicate — keep only 'sent' if both sending+sent exist for same class+ts
const deduped = computed(() => {
  const seen = new Set<string>()
  return sorted.value.filter(n => {
    if (n.status === 'sending') return false // skip sending, keep sent
    const key = `${n.rid}-${n.class}-${n.channel}`
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
})

const paged = computed(() =>
  deduped.value.slice((page.value - 1) * perPage.value, page.value * perPage.value)
)

onMounted(load)

function timeAgo(ts: number): string {
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

const CHANNEL_STYLES: Record<string, string> = {
  mail: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
  database: 'bg-sky-500/10 text-sky-400 border-sky-500/20',
  slack: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
  vonage: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
}
</script>

<template>
  <div class="p-6 space-y-4">
    <h1 class="text-2xl font-bold">Notifications ({{ deduped.length }})</h1>

    <NewItemsBanner @refresh="load" />

    <div v-if="!deduped.length" class="text-zinc-500 py-8 text-center">No notifications captured yet</div>

    <div v-if="deduped.length" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Channel</th>
            <th class="px-4 py-2 text-left">Notification</th>
            <th class="px-4 py-2 text-left">Notifiable</th>
            <th class="px-4 py-2 text-left">Trace</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(n, i) in paged"
            :key="i"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer transition-colors"
            @click="$router.push(`/notifications/${n.rid}/${n.eventIndex}`)"
          >
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs whitespace-nowrap">{{ timeAgo(n.ts) }}</td>
            <td class="px-4 py-2">
              <span class="inline-flex px-1.5 py-0.5 text-xs rounded border"
                :class="CHANNEL_STYLES[n.channel] || 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20'"
              >{{ n.channel }}</span>
            </td>
            <td class="px-4 py-2 font-mono text-xs text-teal-400 truncate max-w-sm">{{ n.class }}</td>
            <td class="px-4 py-2 text-xs text-zinc-400 truncate max-w-[200px]">{{ n.notifiable }}</td>
            <td class="px-4 py-2 text-xs text-zinc-500 font-mono truncate max-w-[200px]">{{ n.traceUri }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="deduped.length" />
  </div>
</template>
