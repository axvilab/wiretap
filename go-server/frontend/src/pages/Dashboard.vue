<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { api, type DashboardData, type Trace } from '@/api'
import { useWebSocket } from '@/hooks/useWebSocket'
import { TraceType, ROOT_TRACE_TYPES } from '@/constants'
import StatusBadge from '@/components/StatusBadge.vue'
import MethodBadge from '@/components/MethodBadge.vue'
import WarningIcons from '@/components/WarningIcons.vue'
import VChart from 'vue-echarts'
import { Activity, Database, Bug, Mail, Briefcase, Clock, Zap } from 'lucide-vue-next'

const period = ref<'1h' | '24h' | '7d'>('1h')
const data = ref<DashboardData | null>(null)
const recent = ref<Trace[]>([])

async function load() {
  const [dash, recentData] = await Promise.all([
    api.dashboard(period.value),
    api.requests(`per_page=10&type=${ROOT_TRACE_TYPES.join(',')}`),
  ])
  data.value = dash
  recent.value = recentData.traces
}

onMounted(load)
watch(period, load)

useWebSocket((event) => {
  if (event.type === 'trace.new' && event.data) {
    if (ROOT_TRACE_TYPES.includes(event.data.type)) {
      recent.value = [event.data, ...recent.value].slice(0, 10)
    }
    // Refresh dashboard data on new traces
    load()
  }
})

// Charts
const timelineChart = computed(() => {
  if (!data.value?.timeline.length) return null
  const t = data.value.timeline
  return {
    tooltip: { trigger: 'axis', backgroundColor: '#18181b', borderColor: '#3f3f46', textStyle: { color: '#e4e4e7' } },
    grid: { top: 10, right: 10, bottom: 24, left: 40 },
    xAxis: { type: 'category', data: t.map(p => p.timestamp), axisLabel: { color: '#71717a', fontSize: 10 }, axisLine: { lineStyle: { color: '#3f3f46' } } },
    yAxis: { type: 'value', axisLabel: { color: '#71717a', fontSize: 10 }, splitLine: { lineStyle: { color: '#27272a' } } },
    series: [
      { name: 'Requests', type: 'bar', data: t.map(p => p.count), itemStyle: { color: '#38bdf8', borderRadius: [2, 2, 0, 0] }, barMaxWidth: 20 },
      { name: 'Errors', type: 'bar', data: t.map(p => p.error_count), itemStyle: { color: '#ef4444', borderRadius: [2, 2, 0, 0] }, barMaxWidth: 20 },
    ],
  }
})

const durationChart = computed(() => {
  if (!data.value?.timeline.length) return null
  const t = data.value.timeline
  return {
    tooltip: { trigger: 'axis', backgroundColor: '#18181b', borderColor: '#3f3f46', textStyle: { color: '#e4e4e7' } },
    grid: { top: 10, right: 10, bottom: 24, left: 50 },
    xAxis: { type: 'category', data: t.map(p => p.timestamp), axisLabel: { color: '#71717a', fontSize: 10 }, axisLine: { lineStyle: { color: '#3f3f46' } } },
    yAxis: { type: 'value', axisLabel: { color: '#71717a', fontSize: 10, formatter: '{value}ms' }, splitLine: { lineStyle: { color: '#27272a' } } },
    series: [
      { name: 'Avg', type: 'line', data: t.map(p => p.avg_ms), smooth: true, lineStyle: { color: '#34d399' }, itemStyle: { color: '#34d399' }, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(52, 211, 153, 0.2)' }, { offset: 1, color: 'rgba(52, 211, 153, 0)' }] } }, symbol: 'none' },
      { name: 'P95', type: 'line', data: t.map(p => p.p95_ms), smooth: true, lineStyle: { color: '#f59e0b', type: 'dashed' }, itemStyle: { color: '#f59e0b' }, symbol: 'none' },
    ],
  }
})

const statusChart = computed(() => {
  if (!data.value?.status_breakdown.length) return null
  const colors: Record<string, string> = { '2xx': '#34d399', '3xx': '#38bdf8', '4xx': '#fbbf24', '5xx': '#ef4444', other: '#71717a' }
  return {
    tooltip: { backgroundColor: '#18181b', borderColor: '#3f3f46', textStyle: { color: '#e4e4e7' } },
    legend: { bottom: 0, textStyle: { color: '#a1a1aa', fontSize: 11 }, itemWidth: 10, itemHeight: 10 },
    series: [{
      type: 'pie',
      radius: ['50%', '80%'],
      center: ['50%', '50%'],
      data: data.value.status_breakdown.map(s => ({
        name: s.group,
        value: s.count,
        itemStyle: { color: colors[s.group] || '#71717a' },
      })),
      label: { show: false },
      emphasis: { label: { fontSize: 13, fontWeight: 'bold' } },
    }],
  }
})

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
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">Dashboard</h1>
      <div class="flex bg-zinc-800 rounded p-0.5">
        <button
          v-for="p in (['1h', '24h', '7d'] as const)"
          :key="p"
          @click="period = p"
          class="px-3 py-1 text-sm rounded transition-colors"
          :class="period === p ? 'bg-zinc-700 text-zinc-100' : 'text-zinc-400 hover:text-zinc-200'"
        >{{ p }}</button>
      </div>
    </div>

    <!-- Stats cards -->
    <!-- Stats row 1: counters (clickable) -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3" v-if="data">
      <router-link to="/requests" class="bg-zinc-900 border border-zinc-800 rounded-lg p-3 hover:border-zinc-700 transition-colors">
        <div class="flex items-center gap-1.5 text-zinc-500 text-xs mb-1"><Activity class="w-3.5 h-3.5" />Requests</div>
        <div class="text-xl font-bold">{{ data.stats.total_requests }}</div>
      </router-link>
      <router-link to="/jobs" class="bg-zinc-900 border border-zinc-800 rounded-lg p-3 hover:border-zinc-700 transition-colors">
        <div class="flex items-center gap-1.5 text-zinc-500 text-xs mb-1"><Briefcase class="w-3.5 h-3.5" />Jobs</div>
        <div class="text-xl font-bold">{{ data.stats.total_jobs }}</div>
      </router-link>
      <router-link to="/exceptions" class="bg-zinc-900 border border-zinc-800 rounded-lg p-3 hover:border-zinc-700 transition-colors">
        <div class="flex items-center gap-1.5 text-zinc-500 text-xs mb-1"><Bug class="w-3.5 h-3.5" />Exceptions</div>
        <div class="text-xl font-bold text-red-400">{{ data.stats.total_exceptions }}</div>
      </router-link>
      <router-link to="/mail" class="bg-zinc-900 border border-zinc-800 rounded-lg p-3 hover:border-zinc-700 transition-colors">
        <div class="flex items-center gap-1.5 text-zinc-500 text-xs mb-1"><Mail class="w-3.5 h-3.5" />Emails</div>
        <div class="text-xl font-bold">{{ data.stats.total_emails }}</div>
      </router-link>
    </div>

    <!-- Stats row 2: performance metrics -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3" v-if="data">
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
        <div class="flex items-center gap-1.5 text-zinc-500 text-xs mb-1"><Clock class="w-3.5 h-3.5" />Avg Response</div>
        <div class="text-xl font-bold">{{ data.stats.avg_duration }}ms</div>
      </div>
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
        <div class="flex items-center gap-1.5 text-zinc-500 text-xs mb-1"><Clock class="w-3.5 h-3.5" />Median</div>
        <div class="text-xl font-bold">{{ data.stats.median_duration }}ms</div>
      </div>
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
        <div class="flex items-center gap-1.5 text-zinc-500 text-xs mb-1"><Database class="w-3.5 h-3.5" />P95</div>
        <div class="text-xl font-bold text-amber-400">{{ data.stats.p95_duration }}ms</div>
      </div>
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
        <div class="flex items-center gap-1.5 text-zinc-500 text-xs mb-1"><Zap class="w-3.5 h-3.5" />Req/min</div>
        <div class="text-xl font-bold text-emerald-400">{{ data.stats.requests_per_min }}</div>
      </div>
    </div>

    <!-- Charts row -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4" v-if="data">
      <!-- Requests timeline -->
      <div class="lg:col-span-2 bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <h3 class="text-sm font-semibold text-zinc-400 mb-3">Requests</h3>
        <VChart v-if="timelineChart" :option="timelineChart" style="height: 200px" autoresize />
        <div v-else class="h-[200px] flex items-center justify-center text-zinc-600 text-sm">No data</div>
      </div>

      <!-- Status breakdown -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <h3 class="text-sm font-semibold text-zinc-400 mb-3">Status codes</h3>
        <VChart v-if="statusChart" :option="statusChart" style="height: 200px" autoresize />
        <div v-else class="h-[200px] flex items-center justify-center text-zinc-600 text-sm">No data</div>
      </div>
    </div>

    <!-- Duration chart -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4" v-if="data">
      <h3 class="text-sm font-semibold text-zinc-400 mb-3">Response time <span class="font-normal text-zinc-600">— avg / p95</span></h3>
      <VChart v-if="durationChart" :option="durationChart" style="height: 180px" autoresize />
      <div v-else class="h-[180px] flex items-center justify-center text-zinc-600 text-sm">No data</div>
    </div>

    <!-- Slow endpoints + Recent Activity -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- Slow endpoints -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg" v-if="data?.slow_endpoints.length">
        <div class="px-4 py-3 border-b border-zinc-800 font-semibold text-sm">Slow Endpoints</div>
        <table class="w-full text-sm">
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

      <!-- Recent Activity -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg">
        <div class="px-4 py-3 border-b border-zinc-800 font-semibold text-sm">Recent Activity</div>
        <table class="w-full text-sm">
          <thead>
            <tr class="text-zinc-500 text-xs border-b border-zinc-800">
              <th class="px-3 py-2 text-left">Time</th>
              <th class="px-3 py-2 text-left">Type</th>
              <th class="px-3 py-2 text-left">Action</th>
              <th class="px-3 py-2 text-right">Duration</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="t in recent"
              :key="t.rid"
              class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer transition-colors"
              @click="$router.push(`/trace/${t.rid}`)"
            >
              <td class="px-3 py-2 text-zinc-500 font-mono text-xs">{{ timeAgo(t.ts_start) }}</td>
              <td class="px-3 py-2">
                <span class="inline-flex px-1.5 py-0.5 text-xs rounded border"
                  :class="t.type === TraceType.REQUEST ? 'bg-sky-500/10 text-sky-400 border-sky-500/20' :
                          t.type === TraceType.COMMAND ? 'bg-orange-500/10 text-orange-400 border-orange-500/20' :
                          t.type === TraceType.SCHEDULE ? 'bg-violet-500/10 text-violet-400 border-violet-500/20' :
                          'bg-zinc-500/10 text-zinc-400 border-zinc-500/20'"
                >{{ t.type }}</span>
                <MethodBadge v-if="t.method" :method="t.method" class="ml-1" />
              </td>
              <td class="px-3 py-2 font-mono text-xs truncate max-w-[200px]">{{ traceLabel(t) }}</td>
              <td class="px-3 py-2 text-right font-mono text-xs">{{ t.duration_ms.toFixed(1) }}ms</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
