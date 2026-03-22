<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
import { useWebSocket } from '@/hooks/useWebSocket'

const health = ref<any>(null)
const { connected } = useWebSocket(() => {})

const tab = ref('general')
const tabs = [
  { id: 'general', label: 'General' },
  { id: 'storage', label: 'Storage' },
  { id: 'info', label: 'Info' },
]
const slowQueryMs = ref('100')
const slowRequestMs = ref('1000')
const retentionDays = ref('7')
const saving = ref(false)
const saveResult = ref<{ ok: boolean; msg: string } | null>(null)

// Storage
const storageStats = ref<any>(null)
const clearDays = ref(0)
const clearing = ref(false)
const clearResult = ref<{ ok: boolean; msg: string } | null>(null)


onMounted(async () => {
  try {
    const res = await fetch('/api/health')
    health.value = await res.json()
  } catch (e: any) {
    health.value = { status: 'error', message: e.message }
  }

  try {
    const settings = await api.getSettings()
    slowQueryMs.value = settings.slow_query_ms || '100'
    slowRequestMs.value = settings.slow_request_ms || '1000'
    retentionDays.value = settings.retention_days || '7'
  } catch {}

  try {
    const res = await fetch('/api/storage/stats')
    storageStats.value = await res.json()
  } catch {}
})

async function loadStorageStats() {
  try {
    const res = await fetch('/api/storage/stats')
    storageStats.value = await res.json()
  } catch {}
}

async function clearData() {
  clearing.value = true
  clearResult.value = null
  try {
    const res = await fetch('/api/storage/clear', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ days: clearDays.value }),
    })
    if (res.ok) {
      clearResult.value = { ok: true, msg: clearDays.value === 0 ? 'All data cleared' : `Data older than ${clearDays.value} days cleared` }
      await loadStorageStats()
    } else {
      clearResult.value = { ok: false, msg: 'Failed to clear data' }
    }
  } catch (e: any) {
    clearResult.value = { ok: false, msg: e.message }
  } finally {
    clearing.value = false
  }
}

async function saveSettings() {
  saving.value = true
  saveResult.value = null
  try {
    await api.updateSettings({
      slow_query_ms: slowQueryMs.value,
      slow_request_ms: slowRequestMs.value,
      retention_days: retentionDays.value,
    })
    saveResult.value = { ok: true, msg: 'Settings saved' }
  } catch (e: any) {
    saveResult.value = { ok: false, msg: e.message }
  } finally {
    saving.value = false
  }
}


</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">Settings</h1>
      <div class="flex items-center gap-3 text-sm">
        <span class="text-zinc-400">API:</span>
        <span :class="health?.status === 'healthy' ? 'text-emerald-400' : 'text-red-400'">{{ health?.status || '...' }}</span>
        <span class="text-zinc-400 ml-2">WS:</span>
        <span :class="connected ? 'text-emerald-400' : 'text-amber-400'">{{ connected ? 'On' : 'Off' }}</span>
      </div>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 border-b border-zinc-800">
      <button
        v-for="t in tabs"
        :key="t.id"
        @click="tab = t.id"
        class="px-4 py-2 text-sm font-medium transition-colors -mb-px"
        :class="tab === t.id
          ? 'text-emerald-400 border-b-2 border-emerald-400'
          : 'text-zinc-500 hover:text-zinc-300'"
      >{{ t.label }}</button>
    </div>

    <!-- General tab -->
    <template v-if="tab === 'general'">
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4">
        <h2 class="font-semibold text-sm">Analysis Thresholds</h2>
        <p class="text-xs text-zinc-500">Configure when Wiretap flags queries and requests as slow.</p>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="text-xs text-zinc-400 mb-1 block">Slow Query (ms)</label>
            <input v-model="slowQueryMs" type="number" min="1" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono" />
          </div>
          <div>
            <label class="text-xs text-zinc-400 mb-1 block">Slow Request (ms)</label>
            <input v-model="slowRequestMs" type="number" min="1" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono" />
          </div>
        </div>
      </div>

      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4">
        <h2 class="font-semibold text-sm">Data Retention</h2>
        <div class="max-w-xs">
          <label class="text-xs text-zinc-400 mb-1 block">Keep data for (days)</label>
          <input v-model="retentionDays" type="number" min="1" max="365" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono" />
          <p class="text-xs text-zinc-600 mt-1">Traces, emails, and jobs older than this will be automatically deleted.</p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <button @click="saveSettings" :disabled="saving" class="px-4 py-1.5 text-sm bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 rounded font-medium transition-colors">
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
        <span v-if="saveResult" class="text-sm" :class="saveResult.ok ? 'text-emerald-400' : 'text-red-400'">{{ saveResult.msg }}</span>
      </div>
    </template>

    <!-- Storage tab -->
    <template v-if="tab === 'storage'">
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4">
        <h2 class="font-semibold text-sm">Database</h2>
        <div v-if="storageStats" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div class="bg-zinc-800 rounded p-3">
            <div class="text-xs text-zinc-500 mb-1">DB Size</div>
            <div class="text-lg font-bold">{{ storageStats.db_size_mb >= 1 ? storageStats.db_size_mb?.toFixed(1) + ' MB' : (storageStats.db_size_bytes / 1024)?.toFixed(1) + ' KB' }}</div>
          </div>
          <div class="bg-zinc-800 rounded p-3">
            <div class="text-xs text-zinc-500 mb-1">Traces</div>
            <div class="text-lg font-bold">{{ storageStats.total_traces?.toLocaleString() }}</div>
          </div>
          <div class="bg-zinc-800 rounded p-3">
            <div class="text-xs text-zinc-500 mb-1">Emails</div>
            <div class="text-lg font-bold">{{ storageStats.total_emails?.toLocaleString() }}</div>
          </div>
          <div class="bg-zinc-800 rounded p-3">
            <div class="text-xs text-zinc-500 mb-1">Jobs</div>
            <div class="text-lg font-bold">{{ storageStats.total_jobs?.toLocaleString() }}</div>
          </div>
        </div>
        <button @click="loadStorageStats" class="px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200 bg-zinc-800 rounded transition-colors">Refresh</button>
      </div>

      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4">
        <h2 class="font-semibold text-sm">Clear Data</h2>
        <div class="flex items-end gap-3">
          <div>
            <label class="text-xs text-zinc-400 mb-1 block">Clear data older than (days)</label>
            <input v-model.number="clearDays" type="number" min="0" class="w-32 bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono" />
            <p class="text-xs text-zinc-600 mt-1">Set to 0 to clear all data</p>
          </div>
          <button
            @click="clearData"
            :disabled="clearing"
            class="px-4 py-1.5 text-sm bg-red-600 hover:bg-red-500 disabled:opacity-50 rounded font-medium transition-colors"
          >{{ clearing ? 'Clearing...' : clearDays === 0 ? 'Clear All' : `Clear older than ${clearDays}d` }}</button>
        </div>
        <span v-if="clearResult" class="text-sm" :class="clearResult.ok ? 'text-emerald-400' : 'text-red-400'">{{ clearResult.msg }}</span>
      </div>
    </template>

    <!-- Info tab -->
    <template v-if="tab === 'info'">
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3">
        <h2 class="font-semibold text-sm">Ports</h2>
        <div class="grid grid-cols-3 gap-4 text-sm">
          <div>
            <div class="text-zinc-400 text-xs">Web UI</div>
            <div class="font-mono">:8080</div>
          </div>
          <div>
            <div class="text-zinc-400 text-xs">Ingest API</div>
            <div class="font-mono">:9090</div>
          </div>
          <div>
            <div class="text-zinc-400 text-xs">SMTP</div>
            <div class="font-mono">:1025</div>
          </div>
        </div>
      </div>

      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3">
        <h2 class="font-semibold text-sm">Wiretap Server (docker-compose)</h2>
        <pre class="text-xs text-zinc-400 bg-zinc-950 rounded p-3 overflow-x-auto">environment:
  - WIRETAP_PASSWORD=your_password
  - WIRETAP_INGEST_KEY=your_ingest_key
  - WIRETAP_APP_URL=http://laravel:8000
  - WIRETAP_REPLAY_SECRET=wt_your_secret
  - WIRETAP_REDIS_ADDR=redis:6379       # optional
  - WIRETAP_RETENTION_DAYS=7</pre>
      </div>

      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3">
        <h2 class="font-semibold text-sm">Laravel .env (HTTP driver)</h2>
        <pre class="text-xs text-zinc-400 bg-zinc-950 rounded p-3 overflow-x-auto">WIRETAP_ENABLED=true
WIRETAP_DRIVER=http
WIRETAP_URL=http://wiretap:9090/ingest
WIRETAP_INGEST_KEY=your_ingest_key
MAIL_HOST=wiretap
MAIL_PORT=1025

# For replay (run: php artisan wiretap:replay-secret)
WIRETAP_REPLAY_ENABLED=true
WIRETAP_REPLAY_SECRET=wt_your_secret</pre>
      </div>

      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3">
        <h2 class="font-semibold text-sm">Laravel .env (Redis driver)</h2>
        <pre class="text-xs text-zinc-400 bg-zinc-950 rounded p-3 overflow-x-auto">WIRETAP_ENABLED=true
WIRETAP_DRIVER=redis
WIRETAP_REDIS_CONNECTION=default
MAIL_HOST=wiretap
MAIL_PORT=1025

# For replay (run: php artisan wiretap:replay-secret)
WIRETAP_REPLAY_ENABLED=true
WIRETAP_REPLAY_SECRET=wt_your_secret</pre>
      </div>
    </template>
  </div>
</template>
