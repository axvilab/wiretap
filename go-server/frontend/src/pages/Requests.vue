<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Trace } from '@/api'
import StatusBadge from '@/components/StatusBadge.vue'
import MethodBadge from '@/components/MethodBadge.vue'
import WarningIcons from '@/components/WarningIcons.vue'

const traces = ref<Trace[]>([])
const total = ref(0)
const page = ref(1)
const perPage = 50
const methodFilter = ref('')
const typeFilter = ref('')

async function load() {
  const params = new URLSearchParams({ page: String(page.value), per_page: String(perPage) })
  if (methodFilter.value) params.set('method', methodFilter.value)
  if (typeFilter.value) params.set('type', typeFilter.value)
  const data = await api.requests(params.toString())
  traces.value = data.traces
  total.value = data.total
}

onMounted(load)
watch([page, methodFilter, typeFilter], load)

function timeAgo(ts: number): string {
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

const totalPages = () => Math.ceil(total.value / perPage)
</script>

<template>
  <div class="p-6 space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">Requests</h1>
      <div class="flex gap-2">
        <select v-model="methodFilter" class="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-sm">
          <option value="">All methods</option>
          <option v-for="m in ['GET','POST','PUT','PATCH','DELETE']" :key="m" :value="m">{{ m }}</option>
        </select>
        <select v-model="typeFilter" class="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-sm">
          <option value="">All types</option>
          <option value="request">Request</option>
          <option value="job">Job</option>
        </select>
      </div>
    </div>

    <div class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Type</th>
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
            <td class="px-4 py-2 text-xs">
              <span :class="t.type === 'job' ? 'text-purple-400' : 'text-sky-400'">{{ t.type }}</span>
            </td>
            <td class="px-4 py-2"><MethodBadge v-if="t.method" :method="t.method" /></td>
            <td class="px-4 py-2 font-mono text-xs truncate max-w-sm">{{ t.uri || t.meta_json }}</td>
            <td class="px-4 py-2"><StatusBadge v-if="t.status" :status="t.status" /></td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.duration_ms.toFixed(1) }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.query_count }}</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.event_count }}</td>
            <td class="px-4 py-2"><WarningIcons :trace="t" /></td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages() > 1" class="flex items-center justify-center gap-2">
      <button
        :disabled="page <= 1"
        @click="page--"
        class="px-3 py-1 text-sm bg-zinc-800 rounded hover:bg-zinc-700 disabled:opacity-40"
      >Prev</button>
      <span class="text-sm text-zinc-400">{{ page }} / {{ totalPages() }}</span>
      <button
        :disabled="page >= totalPages()"
        @click="page++"
        class="px-3 py-1 text-sm bg-zinc-800 rounded hover:bg-zinc-700 disabled:opacity-40"
      >Next</button>
    </div>
  </div>
</template>
