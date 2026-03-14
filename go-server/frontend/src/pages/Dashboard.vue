<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type Trace } from '@/api'
import StatusBadge from '@/components/StatusBadge.vue'
import MethodBadge from '@/components/MethodBadge.vue'
import WarningIcons from '@/components/WarningIcons.vue'
import { Activity, Database, Bug, Mail } from 'lucide-vue-next'

const recent = ref<Trace[]>([])
const stats = ref({ total: 0, exceptions: 0, avgDuration: 0 })
const mailCount = ref(0)

onMounted(async () => {
  const [reqData, mailData] = await Promise.all([
    api.requests('per_page=10'),
    api.mail('per_page=1'),
  ])
  recent.value = reqData.traces
  stats.value.total = reqData.total
  stats.value.exceptions = reqData.traces.filter(t => t.has_exception).length
  stats.value.avgDuration = reqData.traces.length
    ? Math.round(reqData.traces.reduce((s, t) => s + t.duration_ms, 0) / reqData.traces.length)
    : 0
  mailCount.value = mailData.total
})

function timeAgo(ts: number): string {
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}
</script>

<template>
  <div class="p-6 space-y-6">
    <h1 class="text-2xl font-bold">Dashboard</h1>

    <!-- Stats cards -->
    <div class="grid grid-cols-4 gap-4">
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <div class="flex items-center gap-2 text-zinc-400 text-sm mb-1">
          <Activity class="w-4 h-4" /> Total Traces
        </div>
        <div class="text-2xl font-bold">{{ stats.total }}</div>
      </div>
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <div class="flex items-center gap-2 text-zinc-400 text-sm mb-1">
          <Database class="w-4 h-4" /> Avg Duration
        </div>
        <div class="text-2xl font-bold">{{ stats.avgDuration }}ms</div>
      </div>
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <div class="flex items-center gap-2 text-zinc-400 text-sm mb-1">
          <Bug class="w-4 h-4" /> Exceptions
        </div>
        <div class="text-2xl font-bold text-red-400">{{ stats.exceptions }}</div>
      </div>
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <div class="flex items-center gap-2 text-zinc-400 text-sm mb-1">
          <Mail class="w-4 h-4" /> Emails
        </div>
        <div class="text-2xl font-bold">{{ mailCount }}</div>
      </div>
    </div>

    <!-- Recent requests -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <div class="px-4 py-3 border-b border-zinc-800 font-semibold text-sm">Recent Requests</div>
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Method</th>
            <th class="px-4 py-2 text-left">URI</th>
            <th class="px-4 py-2 text-left">Status</th>
            <th class="px-4 py-2 text-right">Duration</th>
            <th class="px-4 py-2 text-right">Queries</th>
            <th class="px-4 py-2 text-left">Warnings</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="t in recent"
            :key="t.rid"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer transition-colors"
            @click="$router.push(`/trace/${t.rid}`)"
          >
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs">{{ timeAgo(t.ts_start) }}</td>
            <td class="px-4 py-2"><MethodBadge :method="t.method" /></td>
            <td class="px-4 py-2 font-mono text-xs truncate max-w-xs">{{ t.uri }}</td>
            <td class="px-4 py-2"><StatusBadge :status="t.status" /></td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.duration_ms.toFixed(1) }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.query_count }}</td>
            <td class="px-4 py-2"><WarningIcons :trace="t" /></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
