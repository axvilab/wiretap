<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { api } from '@/api'

type ReplayType = 'request' | 'job' | 'broadcast' | 'command'

const props = defineProps<{
  show: boolean
  type: ReplayType
  data: Record<string, any>
}>()

const emit = defineEmits<{
  close: []
  success: [result: any]
}>()

// Request fields
const method = ref('GET')
const uri = ref('/')
const headersText = ref('')
const body = ref('')

// Broadcast fields
const eventName = ref('')
const channelsJson = ref('[]')
const payloadJson = ref('{}')

// Command fields
const command = ref('')
const commandInput = ref('')

// Job fields
const jobUuid = ref('')
const jobSerialized = ref('')
const jobPayloadJson = ref('{}')
const jobClass = ref('')
const payloadSize = ref(0)

// State
const loading = ref(false)
const result = ref<{ success: boolean; message: string } | null>(null)

const METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']
const LARGE_PAYLOAD_THRESHOLD = 64 * 1024 // 64KB

watch(() => props.show, (visible) => {
  if (!visible) return
  result.value = null
  loading.value = false

  const d = props.data
  switch (props.type) {
    case 'request':
      method.value = d.method || 'GET'
      uri.value = d.uri || '/'
      headersText.value = d.headers ? JSON.stringify(d.headers, null, 2) : '{}'
      body.value = d.body || ''
      break
    case 'job':
      jobUuid.value = d.uuid || ''
      jobSerialized.value = d.job_serialized || ''
      jobPayloadJson.value = d.job_payload ? JSON.stringify(d.job_payload, null, 2) : '{}'
      jobClass.value = d.job_class || ''
      payloadSize.value = d.payload_size || 0
      break
    case 'broadcast':
      eventName.value = d.event || ''
      channelsJson.value = JSON.stringify(d.channels || [], null, 2)
      payloadJson.value = JSON.stringify(d.payload || {}, null, 2)
      break
    case 'command':
      command.value = d.command || ''
      commandInput.value = d.input || ''
      break
  }
})

const hasJobPayload = computed(() => jobSerialized.value !== '')
const isFailedJob = computed(() => props.data?.status === 'failed')
const isLargePayload = computed(() => payloadSize.value > LARGE_PAYLOAD_THRESHOLD)

const title = computed(() => {
  switch (props.type) {
    case 'request': return 'Replay Request'
    case 'job': return isFailedJob.value && !hasJobPayload.value ? 'Retry Job' : 'Replay Job'
    case 'broadcast': return 'Replay Broadcast'
    case 'command': return 'Replay Command'
  }
})

const hasMaskedFields = computed(() => {
  if (props.type === 'request') return body.value.includes('***')
  if (props.type === 'job') return jobPayloadJson.value.includes('***')
  return false
})

async function send() {
  loading.value = true
  result.value = null
  try {
    let res: any
    switch (props.type) {
      case 'request': {
        let headers: Record<string, string> = {}
        try { headers = JSON.parse(headersText.value) } catch {}
        res = await api.replayRequest({ method: method.value, uri: uri.value, headers, body: body.value })
        result.value = { success: res.success, message: res.success ? `Response: ${res.status}` : res.error }
        break
      }
      case 'job':
        if (hasJobPayload.value) {
          // Dispatch with payload (works for any job)
          let overrides: Record<string, unknown> = {}
          try { overrides = JSON.parse(jobPayloadJson.value) } catch {}
          res = await api.dispatchJob({ job_serialized: jobSerialized.value, job_payload: overrides })
          result.value = { success: res.success, message: res.success ? `Dispatched: ${res.job_class}` : res.error }
        } else {
          // Retry failed job by UUID
          res = await api.retryJob(jobUuid.value)
          result.value = { success: res.success, message: res.success ? (res.output || 'Job queued for retry') : res.error }
        }
        break
      case 'broadcast': {
        let channels: { name: string; type: string }[] = []
        let payload: Record<string, unknown> = {}
        try { channels = JSON.parse(channelsJson.value) } catch {}
        try { payload = JSON.parse(payloadJson.value) } catch {}
        res = await api.replayBroadcast({ event: eventName.value, channels, payload })
        result.value = { success: res.success, message: res.success ? 'Broadcast sent' : res.error }
        break
      }
      case 'command':
        res = await api.replayCommand({ command: command.value, input: commandInput.value })
        result.value = {
          success: res.success,
          message: res.success ? `Exit code: ${res.exit_code}${res.output ? '\n' + res.output : ''}` : res.error,
        }
        break
    }
    if (res?.success) emit('success', res)
  } catch (e: any) {
    result.value = { success: false, message: e.message }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="emit('close')">
      <div class="bg-zinc-900 border border-zinc-700 rounded-lg shadow-xl w-full max-w-lg mx-4">
        <!-- Header -->
        <div class="flex items-center justify-between px-5 py-3 border-b border-zinc-800">
          <h3 class="text-lg font-semibold">{{ title }}</h3>
          <button @click="emit('close')" class="text-zinc-500 hover:text-zinc-300 text-xl leading-none">&times;</button>
        </div>

        <!-- Body -->
        <div class="px-5 py-4 space-y-3 max-h-[60vh] overflow-y-auto">
          <!-- Masked fields warning -->
          <div v-if="hasMaskedFields" class="bg-amber-500/10 border border-amber-500/20 rounded px-3 py-2 text-sm text-amber-400">
            Contains masked fields (***). Edit them before sending.
          </div>

          <!-- Request form -->
          <template v-if="type === 'request'">
            <div class="flex gap-2">
              <select v-model="method" class="bg-zinc-800 border border-zinc-700 rounded px-2 py-1.5 text-sm w-28">
                <option v-for="m in METHODS" :key="m" :value="m">{{ m }}</option>
              </select>
              <input v-model="uri" class="bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm flex-1 font-mono" placeholder="/api/endpoint" />
            </div>
            <div>
              <label class="text-xs text-zinc-500 mb-1 block">Headers (JSON)</label>
              <textarea v-model="headersText" rows="3" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm font-mono resize-y" />
            </div>
            <div>
              <label class="text-xs text-zinc-500 mb-1 block">Body</label>
              <textarea v-model="body" rows="5" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm font-mono resize-y" placeholder="Request body..." />
            </div>
          </template>

          <!-- Job form -->
          <template v-if="type === 'job'">
            <div v-if="isLargePayload" class="bg-amber-500/10 border border-amber-500/20 rounded px-3 py-2 text-sm text-amber-400">
              Large job payload ({{ (payloadSize / 1024).toFixed(1) }}KB). Consider passing IDs instead of full data objects to your jobs.
            </div>

            <template v-if="hasJobPayload">
              <div>
                <label class="text-xs text-zinc-500 mb-1 block">Job Class</label>
                <input :value="jobClass" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono text-purple-400" readonly />
              </div>
              <div>
                <label class="text-xs text-zinc-500 mb-1 block">Properties (JSON)</label>
                <textarea v-model="jobPayloadJson" rows="6" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm font-mono resize-y" />
              </div>
              <p class="text-xs text-zinc-500">Edit public properties above. The job will be dispatched with these values.</p>
            </template>

            <template v-else>
              <div>
                <label class="text-xs text-zinc-500 mb-1 block">Job UUID</label>
                <input v-model="jobUuid" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono" readonly />
              </div>
              <p class="text-sm text-zinc-400">This will retry the failed job via <code class="text-zinc-300">queue:retry</code>.</p>
            </template>
          </template>

          <!-- Broadcast form -->
          <template v-if="type === 'broadcast'">
            <div>
              <label class="text-xs text-zinc-500 mb-1 block">Event Name</label>
              <input v-model="eventName" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono" />
            </div>
            <div>
              <label class="text-xs text-zinc-500 mb-1 block">Channels (JSON)</label>
              <textarea v-model="channelsJson" rows="3" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm font-mono resize-y" />
            </div>
            <div>
              <label class="text-xs text-zinc-500 mb-1 block">Payload (JSON)</label>
              <textarea v-model="payloadJson" rows="5" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm font-mono resize-y" />
            </div>
          </template>

          <!-- Command form -->
          <template v-if="type === 'command'">
            <div>
              <label class="text-xs text-zinc-500 mb-1 block">Command</label>
              <input v-model="command" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono" />
            </div>
            <div>
              <label class="text-xs text-zinc-500 mb-1 block">Arguments</label>
              <input v-model="commandInput" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono" placeholder="--option=value arg1 arg2" />
            </div>
          </template>

          <!-- Result -->
          <div v-if="result" class="rounded px-3 py-2 text-sm whitespace-pre-wrap"
            :class="result.success ? 'bg-emerald-500/10 border border-emerald-500/20 text-emerald-400' : 'bg-red-500/10 border border-red-500/20 text-red-400'"
          >
            {{ result.message }}
          </div>
        </div>

        <!-- Footer -->
        <div class="flex justify-end gap-2 px-5 py-3 border-t border-zinc-800">
          <button @click="emit('close')" class="px-3 py-1.5 text-sm text-zinc-400 hover:text-zinc-200 transition-colors">
            Cancel
          </button>
          <button
            @click="send"
            :disabled="loading"
            class="px-4 py-1.5 text-sm bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 rounded font-medium transition-colors"
          >
            {{ loading ? 'Sending...' : 'Send' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
