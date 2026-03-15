<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
import { EventType } from '@/constants'

const props = defineProps<{ rid: string }>()

const broadcast = ref<any>(null)
const error = ref('')

const CHANNEL_STYLES: Record<string, string> = {
  public: 'bg-sky-500/10 text-sky-400 border-sky-500/20',
  private: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
  presence: 'bg-violet-500/10 text-violet-400 border-violet-500/20',
}

onMounted(async () => {
  try {
    const tree = await api.trace(props.rid)
    const events = JSON.parse(tree.trace.events_json)
    const ev = events.find((e: any) => e.type === EventType.BROADCAST)
    if (ev) {
      broadcast.value = {
        rid: tree.trace.rid,
        event: ev.data.event,
        channels: ev.data.channels || [],
        payload: ev.data.payload || {},
        ts: tree.trace.ts_start + ev.ts,
      }
    } else {
      error.value = 'No broadcast event found in this trace'
    }
  } catch (e: any) {
    error.value = e.message
  }
})

function formatTime(ts: number): string {
  return new Date(ts * 1000).toLocaleString()
}
</script>

<template>
  <div class="p-6 space-y-6">
    <button @click="$router.back()" class="text-zinc-400 hover:text-zinc-100 text-lg">&larr; Back</button>

    <div v-if="error" class="text-red-400">{{ error }}</div>

    <template v-if="broadcast">
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4">
        <!-- Header -->
        <div class="flex items-center justify-between">
          <h1 class="text-xl font-bold font-mono text-indigo-400">{{ broadcast.event }}</h1>
          <span class="text-zinc-500 text-sm">{{ formatTime(broadcast.ts) }}</span>
        </div>

        <!-- Channels -->
        <div>
          <h2 class="text-sm font-semibold text-zinc-400 mb-2">Channels</h2>
          <div class="flex gap-2 flex-wrap">
            <span
              v-for="ch in broadcast.channels"
              :key="ch.name"
              class="inline-flex items-center gap-2 px-3 py-1.5 text-sm rounded border"
              :class="CHANNEL_STYLES[ch.type] || CHANNEL_STYLES.public"
            >
              <span class="opacity-60 text-xs uppercase">{{ ch.type }}</span>
              <span class="font-mono">{{ ch.name }}</span>
            </span>
          </div>
        </div>

        <!-- Payload -->
        <div>
          <h2 class="text-sm font-semibold text-zinc-400 mb-2">Payload</h2>
          <div class="bg-zinc-950 rounded p-4">
            <pre class="text-sm text-zinc-300 whitespace-pre-wrap">{{ JSON.stringify(broadcast.payload, null, 2) }}</pre>
          </div>
        </div>

        <!-- Trace link -->
        <div>
          <router-link
            :to="`/trace/${broadcast.rid}`"
            class="text-sm text-zinc-400 hover:text-zinc-200"
          >&larr; View full trace</router-link>
        </div>
      </div>
    </template>
  </div>
</template>
