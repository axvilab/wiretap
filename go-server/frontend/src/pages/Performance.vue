<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type DashboardData } from '@/api'
import MethodBadge from '@/components/MethodBadge.vue'

const period = ref<string>('1h')
const data = ref<DashboardData | null>(null)

async function load() {
  data.value = await api.dashboard(period.value)
}

onMounted(load)
watch(period, load)
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">Performance</h1>
      <div class="flex bg-zinc-800 rounded p-0.5">
        <button
          v-for="p in ['1m', '5m', '15m', '30m', '1h', '24h', '7d']"
          :key="p"
          @click="period = p"
          class="px-2 py-1 text-xs rounded transition-colors"
          :class="period === p ? 'bg-zinc-700 text-zinc-100' : 'text-zinc-400 hover:text-zinc-200'"
        >{{ p }}</button>
      </div>
    </div>

    <!-- Slow Endpoints -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <div class="px-4 py-3 border-b border-zinc-800 font-semibold text-sm">Slow Endpoints (Top 10)</div>
      <div v-if="!data?.slow_endpoints?.length" class="px-4 py-6 text-zinc-500 text-center text-sm">No slow endpoints</div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Endpoint</th>
            <th class="px-4 py-2 text-right">Hits</th>
            <th class="px-4 py-2 text-right">Avg</th>
            <th class="px-4 py-2 text-right">P95</th>
            <th class="px-4 py-2 text-right">Max</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="e in data.slow_endpoints"
            :key="e.uri"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer transition-colors"
            @click="$router.push(`/requests?uri=${e.uri}`)"
          >
            <td class="px-4 py-2 font-mono text-xs">
              <MethodBadge :method="e.method" />
              <span class="ml-2">{{ e.uri }}</span>
            </td>
            <td class="px-4 py-2 text-right text-xs">{{ e.count }}</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ e.avg_ms }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs text-amber-400">{{ e.p95_ms }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs text-red-400">{{ e.max_ms }}ms</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Slow Queries -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <div class="px-4 py-3 border-b border-zinc-800 font-semibold text-sm">Slow Queries (Top 10)</div>
      <div v-if="!data?.slow_queries?.length" class="px-4 py-6 text-zinc-500 text-center text-sm">No slow queries</div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">SQL</th>
            <th class="px-4 py-2 text-right">Duration</th>
            <th class="px-4 py-2 text-left">Trace</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(q, i) in data.slow_queries"
            :key="i"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer transition-colors"
            @click="$router.push(`/trace/${q.trace_rid}`)"
          >
            <td class="px-4 py-2 font-mono text-xs text-sky-300 truncate max-w-lg">{{ q.sql }}</td>
            <td class="px-4 py-2 text-right font-mono text-xs text-amber-400 whitespace-nowrap">{{ q.time_ms }}ms</td>
            <td class="px-4 py-2 text-xs text-zinc-500 font-mono truncate max-w-[200px]">
              <span class="text-zinc-600">{{ q.trace_type }}</span>
              <span v-if="q.trace_uri" class="ml-1">{{ q.trace_uri }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
