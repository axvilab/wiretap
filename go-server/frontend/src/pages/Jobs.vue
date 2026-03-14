<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type Trace } from '@/api'

const jobs = ref<Trace[]>([])

onMounted(async () => {
  const data = await api.requests('type=job&per_page=100')
  jobs.value = data.traces
})

function getMeta(t: Trace) {
  try { return JSON.parse(t.meta_json) } catch { return {} }
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
    <h1 class="text-2xl font-bold">Jobs</h1>

    <div v-if="!jobs.length" class="text-zinc-500 py-8 text-center">No jobs found</div>

    <div class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <table v-if="jobs.length" class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Class</th>
            <th class="px-4 py-2 text-left">Queue</th>
            <th class="px-4 py-2 text-left">Status</th>
            <th class="px-4 py-2 text-right">Duration</th>
            <th class="px-4 py-2 text-right">Queries</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="j in jobs"
            :key="j.rid"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer"
            @click="$router.push(`/trace/${j.rid}`)"
          >
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs">{{ timeAgo(j.ts_start) }}</td>
            <td class="px-4 py-2 font-mono text-xs text-purple-400">{{ getMeta(j).job_class }}</td>
            <td class="px-4 py-2 text-xs text-zinc-400">{{ getMeta(j).queue || 'default' }}</td>
            <td class="px-4 py-2">
              <span
                class="text-xs px-2 py-0.5 rounded"
                :class="getMeta(j).status === 'failed' ? 'bg-red-500/10 text-red-400' : 'bg-emerald-500/10 text-emerald-400'"
              >{{ getMeta(j).status || 'completed' }}</span>
            </td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ j.duration_ms.toFixed(1) }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ j.query_count }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
