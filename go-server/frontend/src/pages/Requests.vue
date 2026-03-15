<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, type Trace } from '@/api'
import { TraceType } from '@/constants'
import Pagination from '@/components/Pagination.vue'
import NewItemsBanner from '@/components/NewItemsBanner.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import MethodBadge from '@/components/MethodBadge.vue'
import WarningIcons from '@/components/WarningIcons.vue'

const route = useRoute()
const traces = ref<Trace[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(25)
const methodFilter = ref('')
const uriFilter = ref((route.query.uri as string) || '')
const typeFilter = TraceType.REQUEST

async function load() {
  const params = new URLSearchParams({ page: String(page.value), per_page: String(perPage.value) })
  if (methodFilter.value) params.set('method', methodFilter.value)
  if (uriFilter.value) params.set('uri', uriFilter.value)
  params.set('type', typeFilter)
  const data = await api.requests(params.toString())
  traces.value = data.traces
  total.value = data.total
}

onMounted(load)
watch([page, methodFilter, uriFilter, perPage], load)


function timeAgo(ts: number): string {
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

function parseMeta(t: Trace) {
  try { return JSON.parse(t.meta_json) } catch { return {} }
}

</script>

<template>
  <div class="p-6 space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">Requests</h1>
      <div class="flex gap-2">
        <input
          v-model="uriFilter"
          placeholder="Search URI..."
          class="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-sm w-48 placeholder-zinc-500"
        />
        <select v-model="methodFilter" class="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-sm">
          <option value="">All methods</option>
          <option v-for="m in ['GET','POST','PUT','PATCH','DELETE']" :key="m" :value="m">{{ m }}</option>
        </select>
      </div>
    </div>

    <NewItemsBanner :types="[TraceType.REQUEST]" @refresh="load" />

    <div class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Method</th>
            <th class="px-4 py-2 text-left">URI</th>
            <th class="px-4 py-2 text-left">Status</th>
            <th class="px-4 py-2 text-right">Duration</th>
            <th class="px-4 py-2 text-right">Queries</th>
            <th class="px-4 py-2 text-right">Events</th>
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
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs">{{ timeAgo(t.ts_start) }}</td>
            <td class="px-4 py-2"><MethodBadge :method="t.method" /></td>
            <td class="px-4 py-2 font-mono text-xs truncate max-w-sm">{{ t.uri }}</td>
            <td class="px-4 py-2"><StatusBadge v-if="t.status" :status="t.status" /></td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.duration_ms.toFixed(1) }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.query_count }}</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.event_count }}</td>
            <td class="px-4 py-2 whitespace-nowrap"><WarningIcons :trace="t" /></td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="total" />
  </div>
</template>
