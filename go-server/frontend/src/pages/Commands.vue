<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Trace } from '@/api'
import { TraceType } from '@/constants'
import Pagination from '@/components/Pagination.vue'
import ReplayModal from '@/components/ReplayModal.vue'
import NewItemsBanner from '@/components/NewItemsBanner.vue'

const traces = ref<Trace[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(25)
const replayEnabled = ref(false)
const showReplay = ref(false)
const replayData = ref<Record<string, any>>({})

async function load() {
  const data = await api.requests(`type=${TraceType.COMMAND}&page=${page.value}&per_page=${perPage.value}`)
  traces.value = data.traces
  total.value = data.total
}

onMounted(async () => {
  load()
  try {
    const cfg = await api.replayConfig()
    replayEnabled.value = cfg.enabled
  } catch {}
})
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
    <h1 class="text-2xl font-bold">Commands ({{ total }})</h1>

    <NewItemsBanner :types="[TraceType.COMMAND]" @refresh="load" />

    <div v-if="!traces.length" class="text-zinc-500 py-8 text-center">No commands captured yet</div>

    <div v-if="traces.length" class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Command</th>
            <th class="px-4 py-2 text-left">Exit Code</th>
            <th class="px-4 py-2 text-right">Duration</th>
            <th class="px-4 py-2 text-right">Queries</th>
            <th class="px-4 py-2 text-right">Events</th>
            <th v-if="replayEnabled" class="px-4 py-2"></th>
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
            <td class="px-4 py-2 font-mono text-xs text-orange-400">{{ getMeta(t).command }}</td>
            <td class="px-4 py-2">
              <span class="inline-flex px-1.5 py-0.5 text-xs rounded border"
                :class="getMeta(t).exit_code === 0 ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-red-500/10 text-red-400 border-red-500/20'"
              >{{ getMeta(t).exit_code ?? '—' }}</span>
            </td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.duration_ms.toFixed(1) }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.query_count }}</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ t.event_count }}</td>
            <td v-if="replayEnabled" class="px-4 py-2 text-right">
              <button
                @click.stop="replayData = { command: getMeta(t).command, input: getMeta(t).input || '' }; showReplay = true"
                class="px-2 py-0.5 text-xs bg-orange-500/20 text-orange-400 hover:bg-orange-500/30 rounded transition-colors"
              >Replay</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="total" />

    <ReplayModal
      :show="showReplay"
      type="command"
      :data="replayData"
      @close="showReplay = false"
      @success="showReplay = false"
    />
  </div>
</template>
