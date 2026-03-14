<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type TraceNode } from '@/api'
import StatusBadge from '@/components/StatusBadge.vue'
import MethodBadge from '@/components/MethodBadge.vue'

const props = defineProps<{ rid: string }>()
const tree = ref<TraceNode | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    tree.value = await api.trace(props.rid)
  } catch (e: any) {
    error.value = e.message
  }
})

function parseEvents(json: string) {
  try { return JSON.parse(json) } catch { return [] }
}

function parseWarnings(json: string) {
  try { return JSON.parse(json) } catch { return [] }
}

function formatTs(ts: number): string {
  return (ts * 1000).toFixed(1) + 'ms'
}

function eventColor(type: string): string {
  const colors: Record<string, string> = {
    query: 'text-sky-400',
    'request.start': 'text-zinc-400',
    'request.end': 'text-zinc-400',
    exception: 'text-red-400',
    'job.queued': 'text-purple-400',
    'job.start': 'text-purple-400',
    'job.end': 'text-purple-400',
    'cache': 'text-amber-400',
    'cache.write': 'text-amber-400',
    'mail.sending': 'text-emerald-400',
    'mail.sent': 'text-emerald-400',
    transaction: 'text-zinc-500',
  }
  return colors[type] || 'text-zinc-400'
}
</script>

<template>
  <div class="p-6 space-y-6">
    <div v-if="error" class="text-red-400">{{ error }}</div>

    <template v-if="tree">
      <!-- Header -->
      <div class="flex items-center gap-3">
        <button @click="$router.back()" class="text-zinc-400 hover:text-zinc-100">&larr;</button>
        <MethodBadge v-if="tree.trace.method" :method="tree.trace.method" />
        <span class="font-mono text-lg">{{ tree.trace.uri || tree.trace.type }}</span>
        <StatusBadge v-if="tree.trace.status" :status="tree.trace.status" />
        <span class="text-zinc-500 text-sm">{{ tree.trace.duration_ms.toFixed(1) }}ms</span>
      </div>

      <!-- Warnings -->
      <div v-if="parseWarnings(tree.trace.warnings_json).length" class="space-y-1">
        <div
          v-for="(w, i) in parseWarnings(tree.trace.warnings_json)"
          :key="i"
          class="bg-amber-500/10 border border-amber-500/20 rounded px-3 py-2 text-sm"
        >
          <span class="text-amber-400 font-semibold">{{ w.type }}</span>
          <span class="text-zinc-300 ml-2">{{ w.message }}</span>
          <span v-if="w.suggestion" class="text-zinc-500 ml-2">— {{ w.suggestion }}</span>
        </div>
      </div>

      <!-- Events timeline -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg">
        <div class="px-4 py-3 border-b border-zinc-800 font-semibold text-sm">
          Events ({{ tree.trace.event_count }})
        </div>
        <div class="divide-y divide-zinc-800/50">
          <div
            v-for="(ev, i) in parseEvents(tree.trace.events_json)"
            :key="i"
            class="px-4 py-2 flex items-start gap-3 text-sm hover:bg-zinc-800/20"
          >
            <span class="font-mono text-xs text-zinc-500 w-16 shrink-0 text-right">{{ formatTs(ev.ts) }}</span>
            <span class="font-mono text-xs font-bold w-28 shrink-0" :class="eventColor(ev.type)">{{ ev.type }}</span>
            <div class="flex-1 font-mono text-xs text-zinc-300 break-all">
              <template v-if="ev.type === 'query'">
                <span class="text-sky-300">{{ ev.data.sql }}</span>
                <span class="text-zinc-500 ml-2">{{ ev.data.time_ms?.toFixed(2) }}ms</span>
              </template>
              <template v-else-if="ev.type === 'exception'">
                <span class="text-red-300">{{ ev.data.class }}: {{ ev.data.message }}</span>
                <div class="text-zinc-500 text-xs mt-1">{{ ev.data.file }}:{{ ev.data.line }}</div>
              </template>
              <template v-else>
                {{ JSON.stringify(ev.data) }}
              </template>
            </div>
          </div>
        </div>
      </div>

      <!-- Children (job traces) -->
      <div v-if="tree.children.length" class="space-y-4">
        <h2 class="text-lg font-semibold">Child Traces ({{ tree.children.length }})</h2>
        <div
          v-for="child in tree.children"
          :key="child.trace.rid"
          class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 cursor-pointer hover:border-zinc-700"
          @click="$router.push(`/trace/${child.trace.rid}`)"
        >
          <div class="flex items-center gap-2">
            <span class="text-purple-400 text-xs font-bold">{{ child.trace.type }}</span>
            <span class="font-mono text-sm">{{ child.trace.uri || JSON.parse(child.trace.meta_json).job_class }}</span>
            <span class="text-zinc-500 text-sm ml-auto">{{ child.trace.duration_ms.toFixed(1) }}ms</span>
            <span class="text-zinc-600 text-xs">{{ child.trace.query_count }}q {{ child.trace.event_count }}ev</span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
