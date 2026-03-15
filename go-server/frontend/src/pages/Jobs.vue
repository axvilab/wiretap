<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Trace } from '@/api'
import { JobStatus } from '@/constants'
import Pagination from '@/components/Pagination.vue'

const jobs = ref<Trace[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(25)

async function load() {
  const data = await api.requests(`type=job&page=${page.value}&per_page=${perPage.value}`)
  jobs.value = data.traces
  total.value = data.total
}

onMounted(load)
watch([page, perPage], load)
function getMeta(t: Trace) {
  try { return JSON.parse(t.meta_json) } catch { return {} }
}

function timeAgo(ts: number): string {
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

const STATUS_STYLES: Record<string, string> = {
  [JobStatus.COMPLETED]: 'bg-emerald-500/10 text-emerald-400',
  [JobStatus.FAILED]: 'bg-red-500/10 text-red-400',
  [JobStatus.PENDING]: 'bg-zinc-500/10 text-zinc-400',
  [JobStatus.PROCESSING]: 'bg-amber-500/10 text-amber-400',
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
              <span class="text-xs px-2 py-0.5 rounded" :class="STATUS_STYLES[getMeta(j).status] || STATUS_STYLES.completed">
                {{ getMeta(j).status || 'completed' }}
              </span>
            </td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ j.duration_ms.toFixed(1) }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ j.query_count }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="total" />
  </div>
</template>
