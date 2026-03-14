<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type Trace } from '@/api'

const traces = ref<Trace[]>([])

onMounted(async () => {
  // Fetch requests and filter for exceptions client-side for now
  const data = await api.requests('per_page=100')
  traces.value = data.traces.filter(t => t.has_exception)
})

function getException(t: Trace) {
  try {
    const events = JSON.parse(t.events_json)
    const ex = events.find((e: any) => e.type === 'exception')
    return ex?.data || null
  } catch { return null }
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
    <h1 class="text-2xl font-bold">Exceptions</h1>

    <div v-if="!traces.length" class="text-zinc-500 py-8 text-center">No exceptions found</div>

    <div
      v-for="t in traces"
      :key="t.rid"
      class="bg-zinc-900 border border-red-500/20 rounded-lg p-4 cursor-pointer hover:border-red-500/40 transition-colors"
      @click="$router.push(`/trace/${t.rid}`)"
    >
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-2">
          <span class="text-red-400 font-mono text-sm font-bold">{{ getException(t)?.class }}</span>
        </div>
        <span class="text-zinc-500 text-xs">{{ timeAgo(t.ts_start) }}</span>
      </div>
      <div class="text-zinc-300 text-sm">{{ getException(t)?.message }}</div>
      <div class="text-zinc-600 text-xs mt-1 font-mono">
        {{ t.method }} {{ t.uri }} &middot; {{ getException(t)?.file }}:{{ getException(t)?.line }}
      </div>
    </div>
  </div>
</template>
