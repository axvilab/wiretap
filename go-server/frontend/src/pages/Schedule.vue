<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Trace } from '@/api'
import { TraceType } from '@/constants'
import Pagination from '@/components/Pagination.vue'

const traces = ref<Trace[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(25)

async function load() {
  const data = await api.requests(`type=${TraceType.SCHEDULE}&page=${page.value}&per_page=${perPage.value}`)
  traces.value = data.traces
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
</script>

<template>
  <div class="p-6 space-y-4">
    <h1 class="text-2xl font-bold">Schedule ({{ total }})</h1>

    <div v-if="!traces.length" class="text-zinc-500 py-8 text-center">No scheduled tasks captured yet</div>

    <div v-if="traces.length" class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Expression</th>
            <th class="px-4 py-2 text-left">Command</th>
            <th class="px-4 py-2 text-right">Duration</th>
            <th class="px-4 py-2 text-right">Queries</th>
            <th class="px-4 py-2 text-right">Events</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="t in traces"
            :key="t.rid"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer transition-colors"
            @click="$router.push(`/trace/${t.rid}`)"
          >
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs">{{ timeAgo(t.ts_start) }}</td>
            <td class="px-4 py-2">
              <span class="inline-flex px-1.5 py-0.5 text-xs rounded border bg-violet-500/10 text-violet-400 border-violet-500/20 font-mono">
                {{ getMeta(t).expression }}
              </span>
            </td>
            <td class="px-4 py-2 font-mono text-xs text-violet-300">{{ getMeta(t).command }}</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.duration_ms.toFixed(1) }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.query_count }}</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.event_count }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="total" />
  </div>
</template>
