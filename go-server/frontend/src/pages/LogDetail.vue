<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
import { EventType } from '@/constants'

const props = defineProps<{ rid: string; index: string }>()

const log = ref<any>(null)
const error = ref('')

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

onMounted(async () => {
  try {
    const tree = await api.trace(props.rid)
    const events = JSON.parse(tree.trace.events_json)
    const idx = parseInt(props.index)
    const ev = events[idx]

    if (ev && ev.type === EventType.LOG) {
      log.value = {
        rid: tree.trace.rid,
        level: ev.data.level,
        message: ev.data.message,
        context: ev.data.context || {},
        ts: tree.trace.ts_start + ev.ts,
        traceType: tree.trace.type,
        traceUri: tree.trace.uri || tree.trace.type,
      }
    } else {
      error.value = 'Log event not found'
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

    <template v-if="log">
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4">
        <!-- Header -->
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <span class="px-2 py-1 text-sm rounded border font-medium"
              :class="LEVEL_STYLES[log.level] || LEVEL_STYLES.info"
            >{{ log.level }}</span>
            <span class="text-zinc-500 text-sm font-mono">{{ log.traceUri }}</span>
          </div>
          <span class="text-zinc-500 text-sm">{{ formatTime(log.ts) }}</span>
        </div>

        <!-- Message -->
        <div>
          <h2 class="text-sm font-semibold text-zinc-400 mb-2">Message</h2>
          <div class="bg-zinc-950 rounded p-4">
            <pre class="text-sm text-zinc-200 whitespace-pre-wrap break-words">{{ log.message }}</pre>
          </div>
        </div>

        <!-- Context -->
        <div v-if="Object.keys(log.context).length">
          <h2 class="text-sm font-semibold text-zinc-400 mb-2">Context</h2>
          <div class="bg-zinc-950 rounded p-4">
            <pre class="text-sm text-zinc-300 whitespace-pre-wrap break-words">{{ JSON.stringify(log.context, null, 2) }}</pre>
          </div>
        </div>

        <!-- Trace link -->
        <div>
          <router-link
            :to="`/trace/${log.rid}`"
            class="text-sm text-zinc-400 hover:text-zinc-200"
          >&larr; View full trace</router-link>
        </div>
      </div>
    </template>
  </div>
</template>
