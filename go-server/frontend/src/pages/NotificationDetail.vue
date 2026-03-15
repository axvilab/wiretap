<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

const props = defineProps<{ rid: string; index: string }>()

const notification = ref<any>(null)
const error = ref('')

const CHANNEL_STYLES: Record<string, string> = {
  mail: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
  database: 'bg-sky-500/10 text-sky-400 border-sky-500/20',
  slack: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
  vonage: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
}

onMounted(async () => {
  try {
    const tree = await api.trace(props.rid)
    const events = JSON.parse(tree.trace.events_json)
    const idx = parseInt(props.index)
    const ev = events[idx]

    if (ev && (ev.type === 'notification.sent' || ev.type === 'notification.sending')) {
      notification.value = {
        rid: tree.trace.rid,
        channel: ev.data.channel,
        class: ev.data.class,
        notifiable: ev.data.notifiable,
        status: ev.type === 'notification.sent' ? 'sent' : 'sending',
        ts: tree.trace.ts_start + ev.ts,
        traceUri: tree.trace.uri || tree.trace.type,
      }
    } else {
      error.value = 'Notification event not found'
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

    <template v-if="notification">
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4">
        <!-- Header -->
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <span class="px-2 py-1 text-sm rounded border font-medium"
              :class="CHANNEL_STYLES[notification.channel] || 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20'"
            >{{ notification.channel }}</span>
            <span class="text-zinc-500 text-sm font-mono">{{ notification.traceUri }}</span>
          </div>
          <span class="text-zinc-500 text-sm">{{ formatTime(notification.ts) }}</span>
        </div>

        <!-- Class -->
        <div>
          <h2 class="text-sm font-semibold text-zinc-400 mb-2">Notification Class</h2>
          <div class="bg-zinc-950 rounded p-4">
            <pre class="text-sm text-teal-400 font-mono">{{ notification.class }}</pre>
          </div>
        </div>

        <!-- Notifiable -->
        <div>
          <h2 class="text-sm font-semibold text-zinc-400 mb-2">Notifiable</h2>
          <div class="bg-zinc-950 rounded p-4">
            <pre class="text-sm text-zinc-300 font-mono">{{ notification.notifiable }}</pre>
          </div>
        </div>

        <!-- Status -->
        <div>
          <h2 class="text-sm font-semibold text-zinc-400 mb-2">Status</h2>
          <span class="px-2 py-1 text-xs rounded border"
            :class="notification.status === 'sent' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border-amber-500/20'"
          >{{ notification.status }}</span>
        </div>

        <!-- Trace link -->
        <div>
          <router-link
            :to="`/trace/${notification.rid}`"
            class="text-sm text-zinc-400 hover:text-zinc-200"
          >&larr; View full trace</router-link>
        </div>
      </div>
    </template>
  </div>
</template>
