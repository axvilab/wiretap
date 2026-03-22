<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Trace } from '@/api'
import MethodBadge from '@/components/MethodBadge.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import Pagination from '@/components/Pagination.vue'
import NewItemsBanner from '@/components/NewItemsBanner.vue'

const traces = ref<Trace[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(25)

async function load() {
  // Fetch requests with exceptions
  const data = await api.requests(`per_page=100`)
  const all = data.traces.filter(t => t.has_exception)
  total.value = all.length
  traces.value = all.slice((page.value - 1) * perPage.value, page.value * perPage.value)
}

onMounted(load)
watch([page, perPage], load)
function getException(t: Trace) {
  try {
    const events = JSON.parse(t.events_json)
    return events.find((e: any) => e.type === 'exception')?.data || null
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
    <h1 class="text-2xl font-bold">Exceptions ({{ total }})</h1>

    <NewItemsBanner :exceptions-only="true" @refresh="load" />

    <div v-if="!traces.length" class="text-zinc-500 py-8 text-center">No exceptions found</div>

    <div v-if="traces.length" class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Status</th>
            <th class="px-4 py-2 text-left">Class</th>
            <th class="px-4 py-2 text-left">Message</th>
            <th class="px-4 py-2 text-left">Request</th>
            <th class="px-4 py-2 text-left">Location</th>
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
            <td class="px-4 py-2"><StatusBadge v-if="t.status" :status="t.status" /></td>
            <td class="px-4 py-2 font-mono text-xs text-red-400 truncate max-w-xs">{{ getException(t)?.class }}</td>
            <td class="px-4 py-2 text-xs text-zinc-300 truncate max-w-xs">{{ getException(t)?.message }}</td>
            <td class="px-4 py-2 text-xs truncate max-w-[200px]">
              <MethodBadge v-if="t.method" :method="t.method" />
              <span class="ml-1 font-mono text-zinc-400">{{ t.uri }}</span>
            </td>
            <td class="px-4 py-2 font-mono text-xs text-zinc-500 truncate max-w-[200px]">{{ getException(t)?.file }}:{{ getException(t)?.line }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="total" />
  </div>
</template>
