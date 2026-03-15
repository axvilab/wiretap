<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
import { useWebSocket } from '@/hooks/useWebSocket'

const health = ref<any>(null)
const { connected } = useWebSocket(() => {})

const appUrl = ref('')
const replaySecret = ref('')
const ingestApiKey = ref('')
const slowQueryMs = ref('100')
const slowRequestMs = ref('1000')
const saving = ref(false)
const saveResult = ref<{ ok: boolean; msg: string } | null>(null)

onMounted(async () => {
  try {
    const res = await fetch('/api/health')
    health.value = await res.json()
  } catch (e: any) {
    health.value = { status: 'error', message: e.message }
  }

  try {
    const settings = await api.getSettings()
    appUrl.value = settings.app_url || ''
    replaySecret.value = settings.replay_secret || ''
    ingestApiKey.value = settings.ingest_api_key || ''
    slowQueryMs.value = settings.slow_query_ms || '100'
    slowRequestMs.value = settings.slow_request_ms || '1000'
  } catch {}
})

async function saveSettings() {
  saving.value = true
  saveResult.value = null
  try {
    await api.updateSettings({
      app_url: appUrl.value,
      replay_secret: replaySecret.value,
      ingest_api_key: ingestApiKey.value,
      slow_query_ms: slowQueryMs.value,
      slow_request_ms: slowRequestMs.value,
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
    <h1 class="text-2xl font-bold">Settings</h1>

    <!-- Server status -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3">
      <h2 class="font-semibold text-sm">Server Status</h2>
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <span class="text-zinc-400">API:</span>
          <span class="ml-2" :class="health?.status === 'healthy' ? 'text-emerald-400' : 'text-red-400'">
            {{ health?.status || 'checking...' }}
          </span>
        </div>
        <div>
          <span class="text-zinc-400">WebSocket:</span>
          <span class="ml-2" :class="connected ? 'text-emerald-400' : 'text-amber-400'">
            {{ connected ? 'Connected' : 'Disconnected' }}
          </span>
        </div>
      </div>
    </div>

    <!-- Security -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4">
      <h2 class="font-semibold text-sm">Security</h2>
      <div>
        <label class="text-xs text-zinc-400 mb-1 block">Ingest API Key</label>
        <input
          v-model="ingestApiKey"
          type="password"
          class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono"
          placeholder="Leave empty to allow all (not recommended)"
        />
        <p class="text-xs text-zinc-600 mt-1">
          If set, PHP package must send this key via <code class="text-zinc-500">X-Wiretap-Key</code> header.
          Set <code class="text-zinc-500">WIRETAP_INGEST_KEY</code> in your Laravel <code class="text-zinc-500">.env</code>
        </p>
      </div>
    </div>

    <!-- Replay settings -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4">
      <h2 class="font-semibold text-sm">Replay / Retry</h2>
      <p class="text-xs text-zinc-500">
        Configure connection to your Laravel app for replaying requests, retrying jobs, and re-broadcasting events.
      </p>

      <div class="space-y-3">
        <div>
          <label class="text-xs text-zinc-400 mb-1 block">App URL</label>
          <input
            v-model="appUrl"
            class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono"
            placeholder="http://localhost:8000"
          />
          <p class="text-xs text-zinc-600 mt-1">Internal URL of your Laravel app (reachable from Wiretap server)</p>
        </div>
        <div>
          <label class="text-xs text-zinc-400 mb-1 block">Replay Secret</label>
          <input
            v-model="replaySecret"
            type="password"
            class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono"
            placeholder="wt_..."
          />
          <p class="text-xs text-zinc-600 mt-1">
            Must match <code class="text-zinc-500">WIRETAP_REPLAY_SECRET</code> in your Laravel <code class="text-zinc-500">.env</code>.
            Generate with: <code class="text-zinc-500">php artisan wiretap:replay-secret</code>
          </p>
        </div>

      </div>
    </div>

    <!-- Analysis thresholds -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4">
      <h2 class="font-semibold text-sm">Analysis Thresholds</h2>
      <p class="text-xs text-zinc-500">
        Configure when Wiretap flags queries and requests as slow.
      </p>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="text-xs text-zinc-400 mb-1 block">Slow Query (ms)</label>
          <input
            v-model="slowQueryMs"
            type="number"
            min="1"
            class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono"
          />
        </div>
        <div>
          <label class="text-xs text-zinc-400 mb-1 block">Slow Request (ms)</label>
          <input
            v-model="slowRequestMs"
            type="number"
            min="1"
            class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm font-mono"
          />
        </div>
      </div>
    </div>

    <!-- Save button -->
    <div class="flex items-center gap-3">
      <button
        @click="saveSettings"
        :disabled="saving"
        class="px-4 py-1.5 text-sm bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 rounded font-medium transition-colors"
      >{{ saving ? 'Saving...' : 'Save Settings' }}</button>
      <span v-if="saveResult" class="text-sm" :class="saveResult.ok ? 'text-emerald-400' : 'text-red-400'">
        {{ saveResult.msg }}
      </span>
    </div>

    <!-- Ports -->
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

    <!-- PHP config hint -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3">
      <h2 class="font-semibold text-sm">Laravel Configuration</h2>
      <pre class="text-xs text-zinc-400 bg-zinc-950 rounded p-3 overflow-x-auto">WIRETAP_ENABLED=true
WIRETAP_DRIVER=http
WIRETAP_URL=http://wiretap:9090/ingest
MAIL_HOST=wiretap
MAIL_PORT=1025

# For replay (run: php artisan wiretap:replay-secret)
WIRETAP_REPLAY_ENABLED=true
WIRETAP_REPLAY_SECRET=wt_your_secret_here</pre>
    </div>
  </div>
</template>