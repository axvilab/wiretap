<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Trace } from '@/api'
import { useWebSocket } from '@/hooks/useWebSocket'
import { TraceType, ROOT_TRACE_TYPES } from '@/constants'
import StatusBadge from '@/components/StatusBadge.vue'
import MethodBadge from '@/components/MethodBadge.vue'
import WarningIcons from '@/components/WarningIcons.vue'
import Pagination from '@/components/Pagination.vue'
import NewItemsBanner from '@/components/NewItemsBanner.vue'

const traces = ref<Trace[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(25)

async function load() {
  const types = ROOT_TRACE_TYPES.join(',')
  const data = await api.requests(`per_page=${perPage.value}&page=${page.value}&type=${types}`)
  traces.value = data.traces
  total.value = data.total
}

onMounted(load)
watch([page, perPage], load)

function traceLabel(t: Trace): string {
  if (t.uri) return t.uri
  try {
    const meta = JSON.parse(t.meta_json)
    return meta.job_class || meta.command || meta.expression || t.type
  } catch { return t.type }
}

function timeAgo(ts: number): string {
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

const TYPE_STYLES: Record<string, string> = {
  [TraceType.REQUEST]: 'bg-sky-500/10 text-sky-400 border-sky-500/20',
  [TraceType.COMMAND]: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
  [TraceType.SCHEDULE]: 'bg-violet-500/10 text-violet-400 border-violet-500/20',
  [TraceType.JOB]: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
}
</script>

<template>
  <div class="p-6 space-y-4">
    <h1 class="text-2xl font-bold">Activity ({{ total }})</h1>

    <NewItemsBanner :types="[...ROOT_TRACE_TYPES]" @refresh="load" />

    <div v-if="!traces.length" class="text-zinc-500 py-8 text-center">No activity yet</div>

    <div v-if="traces.length" class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Type</th>
            <th class="px-4 py-2 text-left">Action</th>
            <th class="px-4 py-2 text-left">Status</th>
            <th class="px-4 py-2 text-right">Duration</th>
            <th class="px-4 py-2 text-right">Queries</th>
            <th class="px-4 py-2 text-left">Warnings</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="t in traces"
            :key="t.rid"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer transition-colors"
            @click="$router.push(`/trace/${t.rid}`)"
          >
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs whitespace-nowrap">{{ timeAgo(t.ts_start) }}</td>
            <td class="px-4 py-2">
              <span class="inline-flex px-1.5 py-0.5 text-xs rounded border"
                :class="TYPE_STYLES[t.type] || 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20'"
              >{{ t.type }}</span>
              <MethodBadge v-if="t.method" :method="t.method" class="ml-1" />
            </td>
            <td class="px-4 py-2 font-mono text-xs truncate max-w-sm">{{ traceLabel(t) }}</td>
            <td class="px-4 py-2"><StatusBadge v-if="t.status" :status="t.status" /></td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.duration_ms.toFixed(1) }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.query_count }}</td>
            <td class="px-4 py-2"><WarningIcons :trace="t" /></td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="total" />
  </div>
</template>
