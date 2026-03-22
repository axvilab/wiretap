<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Trace } from '@/api'
import { JobStatus, TraceType } from '@/constants'
import Pagination from '@/components/Pagination.vue'
import ReplayModal from '@/components/ReplayModal.vue'
import NewItemsBanner from '@/components/NewItemsBanner.vue'

const jobs = ref<Trace[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(25)
const replayEnabled = ref(false)
const showReplay = ref(false)
const replayJobData = ref<Record<string, any>>({})

async function load() {
  const data = await api.requests(`type=job&page=${page.value}&per_page=${perPage.value}`)
  jobs.value = data.traces
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

function hasPayload(job: Trace): boolean {
  try {
    const events = JSON.parse(job.events_json)
    const startEvent = events.find((ev: any) => ev.type === 'job.start')
    return !!startEvent?.data?.job_serialized
  } catch { return false }
}

function openReplay(job: Trace, e: Event) {
  e.stopPropagation()
  const meta = getMeta(job)
  // Get job_payload from events
  let jobPayload: Record<string, any> = {}
  let jobSerialized = ''
  let payloadSize = 0
  try {
    const events = JSON.parse(job.events_json)
    const startEvent = events.find((ev: any) => ev.type === 'job.start')
    if (startEvent?.data) {
      jobPayload = startEvent.data.job_payload || {}
      jobSerialized = startEvent.data.job_serialized || ''
      payloadSize = startEvent.data.payload_size || 0
    }
  } catch {}

  replayJobData.value = {
    uuid: meta.job_uuid,
    job_class: meta.job_class,
    status: meta.status,
    job_payload: jobPayload,
    job_serialized: jobSerialized,
    payload_size: payloadSize,
  }
  showReplay.value = true
}

function timeAgo(ts: number): string {
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

const STATUS_STYLES: Record<string, string> = {
  [JobStatus.COMPLETED]: 'bg-emerald-500/10 text-emerald-400',
  [JobStatus.FAILED]: 'bg-red-500/10 text-red-400',
  [JobStatus.PENDING]: 'bg-zinc-500/10 text-zinc-400',
  [JobStatus.PROCESSING]: 'bg-amber-500/10 text-amber-400',
}

</script>

<template>
  <div class="p-6 space-y-4">
    <h1 class="text-2xl font-bold">Jobs ({{ total }})</h1>

    <NewItemsBanner :types="[TraceType.JOB]" @refresh="load" />

    <div v-if="!jobs.length" class="text-zinc-500 py-8 text-center">No jobs found</div>

    <div class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
      <table v-if="jobs.length" class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Class</th>
            <th class="px-4 py-2 text-left">Queue</th>
            <th class="px-4 py-2 text-left">Status</th>
            <th class="px-4 py-2 text-right">Duration</th>
            <th class="px-4 py-2 text-right">Queries</th>
            <th v-if="replayEnabled" class="px-4 py-2"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="j in jobs"
            :key="j.rid"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer"
            @click="$router.push(`/trace/${j.rid}`)"
          >
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs">{{ timeAgo(j.ts_start) }}</td>
            <td class="px-4 py-2 font-mono text-xs text-purple-400 truncate max-w-sm">{{ getMeta(j).job_class }}</td>
            <td class="px-4 py-2 text-xs text-zinc-400">{{ getMeta(j).queue || 'default' }}</td>
            <td class="px-4 py-2">
              <span class="text-xs px-2 py-0.5 rounded" :class="STATUS_STYLES[getMeta(j).status] || STATUS_STYLES.completed">
                {{ getMeta(j).status || 'completed' }}
              </span>
            </td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ j.duration_ms.toFixed(1) }}ms</td>
            <td class="px-4 py-2 text-right font-mono text-xs">{{ j.query_count }}</td>
            <td v-if="replayEnabled" class="px-4 py-2 text-right">
              <button
                v-if="getMeta(j).status === 'failed' || hasPayload(j)"
                @click="openReplay(j, $event)"
                class="px-2 py-0.5 text-xs rounded transition-colors"
                :class="getMeta(j).status === 'failed' ? 'bg-red-500/20 text-red-400 hover:bg-red-500/30' : 'bg-indigo-500/20 text-indigo-400 hover:bg-indigo-500/30'"
              >{{ getMeta(j).status === 'failed' && !hasPayload(j) ? 'Retry' : 'Replay' }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="total" />

    <ReplayModal
      :show="showReplay"
      type="job"
      :data="replayJobData"
      @close="showReplay = false"
      @success="showReplay = false; load()"
    />
  </div>
</template>
