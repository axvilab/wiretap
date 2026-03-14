<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useWebSocket } from '@/hooks/useWebSocket'

const health = ref<any>(null)
const { connected } = useWebSocket(() => {})

onMounted(async () => {
  try {
    const res = await fetch('/api/health')
    health.value = await res.json()
  } catch (e: any) {
    health.value = { status: 'error', message: e.message }
  }
})
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
MAIL_PORT=1025</pre>
    </div>
  </div>
</template>
