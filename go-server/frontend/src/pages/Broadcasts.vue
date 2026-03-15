<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Trace } from '@/api'
import { EventType } from '@/constants'
import Pagination from '@/components/Pagination.vue'
import ReplayModal from '@/components/ReplayModal.vue'
import NewItemsBanner from '@/components/NewItemsBanner.vue'

interface BroadcastChannel {
  name: string
  type: string
}

interface BroadcastRow {
  rid: string
  ts: number
  event: string
  channels: BroadcastChannel[]
  payload: Record<string, unknown>
}

const broadcasts = ref<BroadcastRow[]>([])
const page = ref(1)
const perPage = ref(25)
const replayEnabled = ref(false)
const showReplay = ref(false)
const replayData = ref<Record<string, any>>({})

async function load() {
  broadcasts.value = []
  const data = await api.requests('per_page=200')
  for (const t of data.traces) {
    extractBroadcasts(t)
  }
}

onMounted(async () => {
  await load()
  try {
    const cfg = await api.replayConfig()
    replayEnabled.value = cfg.enabled
  } catch {}
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
          payload: ev.data.payload || {},
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

    <NewItemsBanner :broadcast-only="true" @refresh="load" />

    <div v-if="!broadcasts.length" class="text-zinc-500 py-8 text-center">No broadcasts captured yet</div>

    <div v-if="broadcasts.length" class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Event</th>
            <th class="px-4 py-2 text-left">Channels</th>
            <th v-if="replayEnabled" class="px-4 py-2"></th>
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
            <td v-if="replayEnabled" class="px-4 py-2 text-right">
              <button
                @click.stop="replayData = { event: b.event, channels: b.channels, payload: b.payload }; showReplay = true"
                class="px-2 py-0.5 text-xs bg-indigo-500/20 text-indigo-400 hover:bg-indigo-500/30 rounded transition-colors"
              >Replay</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="broadcasts.length" />

    <ReplayModal
      :show="showReplay"
      type="broadcast"
      :data="replayData"
      @close="showReplay = false"
      @success="showReplay = false"
    />
  </div>
</template>
