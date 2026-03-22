<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { api } from '@/api'
import type { MonitorStatus, MonitorSite, NotificationChannel, AlertRule, ScheduleHeartbeat } from '@/api'
import { Plus, Trash2, RefreshCw, Shield, Globe, Heart, Clock, Send, Pencil } from 'lucide-vue-next'

const tab = ref<'sites' | 'notifications' | 'schedules'>('sites')

// --- Sites ---
const statuses = ref<MonitorStatus[]>([])
const showAddSite = ref(false)
const newSite = ref({ url: '', name: '', check_ssl: true, health_url: '', auth_type: 'none', auth_json: '{}', interval_s: 30 })
const newAuth = ref({ token: '', username: '', password: '', header: '', value: '', param: '' })

async function loadStatus() {
  statuses.value = await api.monitorStatus()
}

const authTypes = [
  { value: 'none', label: 'No Auth' },
  { value: 'bearer', label: 'Bearer Token' },
  { value: 'basic', label: 'Basic Auth' },
  { value: 'header', label: 'Custom Header' },
  { value: 'query', label: 'Query Parameter' },
]

function buildAuthJSON(auth: typeof newAuth.value, type: string): string {
  if (type === 'none') return '{}'
  return JSON.stringify(auth)
}

const formError = ref('')

async function addSite() {
  formError.value = ''
  const url = newSite.value.url.trim()
  if (!url) { formError.value = 'URL is required'; return }
  try { const u = new URL(url); if (u.protocol !== 'http:' && u.protocol !== 'https:') throw 0 }
  catch { formError.value = 'Enter a valid http/https URL'; return }
  if (newSite.value.interval_s < 5) { formError.value = 'Interval must be at least 5 seconds'; return }
  try {
    await api.monitorCreateSite({ ...newSite.value, url, auth_json: buildAuthJSON(newAuth.value, newSite.value.auth_type) })
    newSite.value = { url: '', name: '', check_ssl: true, health_url: '', auth_type: 'none', auth_json: '{}', interval_s: 30 }
    newAuth.value = { token: '', username: '', password: '', header: '', value: '', param: '' }
    showAddSite.value = false
    loadStatus()
  } catch (e: any) { formError.value = e.message }
}

async function removeSite(id: number) {
  await api.monitorDeleteSite(id)
  loadStatus()
}

const editingSite = ref<MonitorSite | null>(null)
const editAuth = ref({ token: '', username: '', password: '', header: '', value: '', param: '' })
const editError = ref('')

function startEditSite(s: MonitorStatus) {
  editingSite.value = { ...s.site }
  try { editAuth.value = { token: '', username: '', password: '', header: '', value: '', param: '', ...JSON.parse(s.site.auth_json || '{}') } }
  catch { editAuth.value = { token: '', username: '', password: '', header: '', value: '', param: '' } }
  editError.value = ''
}

async function saveEditSite() {
  if (!editingSite.value) return
  editError.value = ''
  const url = editingSite.value.url.trim()
  if (!url) { editError.value = 'URL is required'; return }
  try { const u = new URL(url); if (u.protocol !== 'http:' && u.protocol !== 'https:') throw 0 }
  catch { editError.value = 'Enter a valid http/https URL'; return }
  try {
    await api.monitorUpdateSite(editingSite.value.id, { ...editingSite.value, url, auth_json: buildAuthJSON(editAuth.value, editingSite.value.auth_type) })
    editingSite.value = null
    loadStatus()
  } catch (e: any) { editError.value = e.message }
}

function formatDate(ts: number) {
  if (!ts) return '—'
  return new Date(ts * 1000).toLocaleDateString()
}

interface HealthCheck {
  name: string
  label: string
  status: string
  shortSummary: string
}

function parseHealthChecks(dataJson: string): HealthCheck[] {
  try {
    const data = JSON.parse(dataJson)
    // Spatie format: { checkResults: [...] }
    if (data.checkResults && Array.isArray(data.checkResults)) {
      return data.checkResults
    }
    // Generic: array of checks
    if (Array.isArray(data)) {
      return data
    }
    // Simple { status: 'ok' } — no details
    return []
  } catch {
    return []
  }
}

function statusColor(s: MonitorStatus) {
  if (!s.last_check) return 'text-zinc-500'
  if (s.last_check.status >= 200 && s.last_check.status < 400) return 'text-emerald-400'
  if (s.last_check.status >= 400) return 'text-red-400'
  return 'text-red-400'
}

function statusLabel(s: MonitorStatus) {
  if (!s.last_check) return 'Pending'
  if (s.last_check.status >= 200 && s.last_check.status < 400) return 'Up'
  return 'Down'
}

// --- Notifications ---
const channels = ref<NotificationChannel[]>([])
const rules = ref<AlertRule[]>([])
const showAddChannel = ref(false)
const newChannel = ref({ type: 'telegram', name: '', config: '', enabled: true })
const channelConfig = ref({ token: '', chat_id: '', webhook_url: '' })
const showAddRule = ref(false)
const newRule = ref({ channel_id: 0, cooldown_m: 5, types: {} as Record<string, boolean> })

function initRuleTypes() {
  const t: Record<string, boolean> = {}
  ruleTypes.forEach(r => t[r] = true)
  return t
}

const ruleTypes = ['exception', 'site_down', 'ssl_expiry', 'domain_expiry', 'schedule_missed']
const channelTypes = ['telegram', 'slack', 'teams']

async function loadNotifications() {
  channels.value = await api.notificationChannels()
  rules.value = await api.alertRules()
}

const channelError = ref('')

async function addChannel() {
  channelError.value = ''
  const t = newChannel.value.type
  if (t === 'telegram') {
    if (!channelConfig.value.token.trim()) { channelError.value = 'Bot token is required'; return }
    if (!channelConfig.value.chat_id.trim()) { channelError.value = 'Chat ID is required'; return }
    newChannel.value.config = JSON.stringify({ token: channelConfig.value.token.trim(), chat_id: channelConfig.value.chat_id.trim() })
  } else {
    if (!channelConfig.value.webhook_url.trim()) { channelError.value = 'Webhook URL is required'; return }
    newChannel.value.config = JSON.stringify({ webhook_url: channelConfig.value.webhook_url.trim() })
  }
  if (!newChannel.value.name.trim()) newChannel.value.name = t
  try {
    await api.createNotificationChannel(newChannel.value)
    newChannel.value = { type: 'telegram', name: '', config: '', enabled: true }
    channelConfig.value = { token: '', chat_id: '', webhook_url: '' }
    showAddChannel.value = false
    loadNotifications()
  } catch (e: any) { channelError.value = e.message }
}

async function removeChannel(id: number) {
  await api.deleteNotificationChannel(id)
  loadNotifications()
}

async function testChannel(id: number) {
  try {
    await api.testNotificationChannel(id)
    alert('Test sent!')
  } catch (e: any) {
    alert('Failed: ' + e.message)
  }
}

const ruleError = ref('')

async function addRule() {
  ruleError.value = ''
  if (!newRule.value.channel_id) { ruleError.value = 'Select a channel'; return }
  if (newRule.value.cooldown_m < 1) { ruleError.value = 'Min interval must be at least 1 minute'; return }
  const selected = Object.entries(newRule.value.types).filter(([, v]) => v).map(([k]) => k)
  if (!selected.length) { ruleError.value = 'Select at least one event type'; return }
  try {
    for (const type of selected) {
      // Skip if rule already exists for this channel+type
      if (rules.value.some(r => r.channel_id === newRule.value.channel_id && r.type === type)) continue
      await api.createAlertRule({ type, channel_id: newRule.value.channel_id, cooldown_m: newRule.value.cooldown_m, enabled: true })
    }
    newRule.value = { channel_id: 0, cooldown_m: 5, types: initRuleTypes() }
    showAddRule.value = false
    loadNotifications()
  } catch (e: any) { ruleError.value = e.message }
}

async function removeRule(id: number) {
  await api.deleteAlertRule(id)
  loadNotifications()
}

function channelName(id: number) {
  return channels.value.find(c => c.id === id)?.name || `#${id}`
}

// --- Schedule Heartbeats ---
const heartbeats = ref<ScheduleHeartbeat[]>([])
const showAddHeartbeat = ref(false)
const newHeartbeat = ref({ command: '', mode: 'every' as 'every' | 'at', interval_h: 0, interval_m: 60, at_time: '00:00', grace_m: 15 })

const knownCommands = ref<string[]>([])

async function loadHeartbeats() {
  heartbeats.value = await api.scheduleHeartbeats()
  knownCommands.value = await api.knownScheduleCommands()
}

const commandUnknown = ref(false)
function checkCommand() {
  const cmd = newHeartbeat.value.command.trim()
  commandUnknown.value = cmd !== '' && !knownCommands.value.includes(cmd)
}

const heartbeatError = ref('')

async function addHeartbeat() {
  heartbeatError.value = ''
  if (!newHeartbeat.value.command.trim()) { heartbeatError.value = 'Command is required'; return }
  if (newHeartbeat.value.grace_m < 1) { heartbeatError.value = 'Grace period must be at least 1 minute'; return }
  const mode = newHeartbeat.value.mode
  let max_gap_m = 0
  let at_time = ''
  if (mode === 'every') {
    const totalInterval = (newHeartbeat.value.interval_h || 0) * 60 + (newHeartbeat.value.interval_m || 0)
    if (totalInterval < 1) { heartbeatError.value = 'Interval must be at least 1 minute'; return }
    max_gap_m = totalInterval + newHeartbeat.value.grace_m
  } else {
    if (!newHeartbeat.value.at_time) { heartbeatError.value = 'Time is required'; return }
    at_time = newHeartbeat.value.at_time
    max_gap_m = 0 // not used in "at" mode
  }
  try {
    await api.createScheduleHeartbeat({ command: newHeartbeat.value.command, mode, max_gap_m, at_time, grace_m: newHeartbeat.value.grace_m })
    newHeartbeat.value = { command: '', mode: 'every', interval_h: 0, interval_m: 60, at_time: '00:00', grace_m: 15 }
    showAddHeartbeat.value = false
    loadHeartbeats()
  } catch (e: any) { heartbeatError.value = e.message }
}

async function removeHeartbeat(id: number) {
  await api.deleteScheduleHeartbeat(id)
  loadHeartbeats()
}

// WebSocket for realtime monitor updates
let ws: WebSocket | null = null

function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  ws = new WebSocket(`${proto}://${location.host}/api/ws`)
  ws.onmessage = (e) => {
    try {
      const msg = JSON.parse(e.data)
      if (msg.type === 'monitor.check' || msg.type === 'monitor.health') {
        loadStatus()
      }
    } catch {}
  }
  ws.onclose = () => { setTimeout(connectWS, 3000) }
}

// Auto-refresh every 30s as fallback
let refreshTimer: ReturnType<typeof setInterval>

onMounted(() => {
  loadStatus()
  loadNotifications()
  loadHeartbeats()
  connectWS()
  refreshTimer = setInterval(loadStatus, 30000)
})

onUnmounted(() => {
  ws?.close()
  ws = null
  clearInterval(refreshTimer)
})
</script>

<template>
  <div class="p-6 space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-lg font-semibold text-zinc-100">Monitoring</h1>
      <button @click="loadStatus(); loadNotifications(); loadHeartbeats()" class="p-1.5 text-zinc-400 hover:text-zinc-200">
        <RefreshCw class="w-4 h-4" />
      </button>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 bg-zinc-900 rounded-lg p-1 w-fit">
      <button v-for="t in (['sites', 'notifications', 'schedules'] as const)" :key="t"
        @click="tab = t"
        class="px-3 py-1.5 text-xs rounded-md transition-colors"
        :class="tab === t ? 'bg-zinc-700 text-zinc-100' : 'text-zinc-400 hover:text-zinc-200'"
      >{{ t === 'sites' ? 'Sites' : t === 'notifications' ? 'Notifications' : 'Schedules' }}</button>
    </div>

    <!-- Sites Tab -->
    <template v-if="tab === 'sites'">
      <div class="space-y-3">
        <!-- Status cards -->
        <div v-for="s in statuses" :key="s.site.id" class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <!-- Edit mode -->
          <template v-if="editingSite?.id === s.site.id">
            <div class="space-y-3 max-w-xl">
              <div class="space-y-2">
                <input v-model="editingSite.name" placeholder="Name (optional)" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200" />
                <input v-model="editingSite.url" placeholder="https://example.com" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200" />
                <input v-model="editingSite.health_url" placeholder="Health URL, optional (e.g. https://api.example.com/health)" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
              </div>
              <div class="flex items-center gap-5 text-xs text-zinc-400">
                <label v-if="editingSite.url.startsWith('https')" class="flex items-center gap-2 cursor-pointer"><input type="checkbox" v-model="editingSite.check_ssl" class="rounded border-zinc-600 bg-zinc-800 text-blue-500" /> Check SSL</label>
                <label class="flex items-center gap-2">Interval: <input type="number" v-model.number="editingSite.interval_s" min="5" class="w-16 bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-zinc-200 text-center" /> s</label>
              </div>
              <div class="space-y-2">
                <select v-model="editingSite.auth_type" class="bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 w-48">
                  <option v-for="a in authTypes" :key="a.value" :value="a.value">{{ a.label }}</option>
                </select>
                <div v-if="editingSite.auth_type === 'bearer'" class="flex gap-2">
                  <input v-model="editAuth.token" placeholder="Token" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
                </div>
                <div v-if="editingSite.auth_type === 'basic'" class="flex gap-2">
                  <input v-model="editAuth.username" placeholder="Username" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
                  <input v-model="editAuth.password" type="password" placeholder="Password" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
                </div>
                <div v-if="editingSite.auth_type === 'header'" class="flex gap-2">
                  <input v-model="editAuth.header" placeholder="Header name" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
                  <input v-model="editAuth.value" placeholder="Header value" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
                </div>
                <div v-if="editingSite.auth_type === 'query'" class="flex gap-2">
                  <input v-model="editAuth.param" placeholder="Param name" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
                  <input v-model="editAuth.value" placeholder="Param value" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
                </div>
              </div>
              <p v-if="editError" class="text-xs text-red-400">{{ editError }}</p>
              <div class="flex gap-2">
                <button @click="saveEditSite" class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-500 transition-colors">Save</button>
                <button @click="editingSite = null" class="px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200 transition-colors">Cancel</button>
              </div>
            </div>
          </template>
          <!-- View mode -->
          <template v-else>
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center gap-3">
              <span class="w-2.5 h-2.5 rounded-full" :class="statusColor(s) === 'text-emerald-400' ? 'bg-emerald-400' : statusColor(s) === 'text-red-400' ? 'bg-red-400' : 'bg-zinc-500'"></span>
              <div>
                <span class="text-zinc-100 font-medium">{{ s.site.name || s.site.url }}</span>
                <span v-if="s.site.name" class="ml-2 text-zinc-500 text-xs">{{ s.site.url }}</span>
                <span v-if="s.site.is_primary" class="ml-2 px-1.5 py-0.5 text-xs bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded">Primary</span>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <span class="px-2 py-0.5 text-xs font-medium rounded-full" :class="statusLabel(s) === 'Up' ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/25' : statusLabel(s) === 'Down' ? 'bg-red-500/15 text-red-400 border border-red-500/25' : 'bg-zinc-500/15 text-zinc-400 border border-zinc-500/25'">{{ statusLabel(s) }}</span>
              <button @click="startEditSite(s)" class="text-zinc-600 hover:text-zinc-300"><Pencil class="w-3.5 h-3.5" /></button>
              <button @click="removeSite(s.site.id)" class="text-zinc-600 hover:text-red-400"><Trash2 class="w-3.5 h-3.5" /></button>
            </div>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
            <!-- Uptime -->
            <div class="bg-zinc-800/50 rounded p-2">
              <div class="text-zinc-500 mb-1">Uptime (24h)</div>
              <div class="text-zinc-200 font-mono">{{ s.uptime_pct.toFixed(1) }}%</div>
            </div>
            <!-- Response time -->
            <div class="bg-zinc-800/50 rounded p-2">
              <div class="text-zinc-500 mb-1">Response</div>
              <div class="text-zinc-200 font-mono">{{ s.last_check?.response_ms || 0 }}ms</div>
            </div>
            <!-- SSL -->
            <div v-if="s.ssl" class="bg-zinc-800/50 rounded p-2">
              <div class="text-zinc-500 mb-1 flex items-center gap-1"><Shield class="w-3 h-3" /> SSL</div>
              <div :class="s.ssl.days_remaining <= 14 ? 'text-amber-400' : s.ssl.days_remaining <= 0 ? 'text-red-400' : 'text-emerald-400'" class="font-mono">
                {{ s.ssl.days_remaining }}d left
              </div>
            </div>
            <!-- Domain -->
            <div v-if="s.domain && s.domain.expires_at" class="bg-zinc-800/50 rounded p-2">
              <div class="text-zinc-500 mb-1 flex items-center gap-1"><Globe class="w-3 h-3" /> Domain</div>
              <div :class="s.domain.days_remaining <= 30 ? 'text-amber-400' : 'text-zinc-200'" class="font-mono">
                {{ s.domain.days_remaining }}d left
              </div>
            </div>
            <!-- Health -->
          </div>

          <!-- Health details -->
          <div v-if="s.site.health_url && s.health" class="mt-3 border-t border-zinc-800 pt-3">
            <div class="flex items-center gap-2 mb-2">
              <Heart class="w-3.5 h-3.5 text-zinc-500" />
              <span class="text-xs font-medium text-zinc-400">Health Checks</span>
              <span class="px-1.5 py-0.5 text-xs rounded-full" :class="s.health.status === 'healthy' ? 'bg-emerald-500/15 text-emerald-400' : 'bg-red-500/15 text-red-400'">{{ s.health.status }}</span>
            </div>
            <div class="flex flex-wrap gap-2">
              <template v-for="check in parseHealthChecks(s.health.data_json)" :key="check.name">
                <div class="flex items-center gap-1.5 px-2 py-1 rounded bg-zinc-800/70 text-xs">
                  <span class="w-1.5 h-1.5 rounded-full" :class="check.status === 'ok' ? 'bg-emerald-400' : check.status === 'warning' ? 'bg-amber-400' : 'bg-red-400'"></span>
                  <span class="text-zinc-300">{{ check.label || check.name }}</span>
                  <span v-if="check.shortSummary && check.shortSummary !== 'Ok'" class="text-zinc-500">{{ check.shortSummary }}</span>
                </div>
              </template>
              <span v-if="!parseHealthChecks(s.health.data_json).length" class="text-xs text-zinc-600">No check details</span>
            </div>
          </div>
          </template>
        </div>

        <!-- Add site -->
        <div v-if="showAddSite" class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3 max-w-xl">
          <h3 class="text-sm font-medium text-zinc-300">Add Site</h3>
          <div class="space-y-2">
            <input v-model="newSite.name" placeholder="Name (optional)" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
            <input v-model="newSite.url" placeholder="https://example.com" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
            <input v-model="newSite.health_url" placeholder="Health URL, optional (e.g. https://api.example.com/health)" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
          </div>
          <div class="flex items-center gap-5 text-xs text-zinc-400">
            <label v-if="newSite.url.startsWith('https')" class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="newSite.check_ssl" class="rounded border-zinc-600 bg-zinc-800 text-blue-500 focus:ring-blue-500/20" />
              <span>Check SSL</span>
            </label>
            <label class="flex items-center gap-2">
              <span>Interval:</span>
              <input type="number" v-model.number="newSite.interval_s" min="5" class="w-16 bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-zinc-200 text-center" />
              <span>s</span>
            </label>
          </div>
          <div class="space-y-2">
            <select v-model="newSite.auth_type" class="bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 w-48">
              <option v-for="a in authTypes" :key="a.value" :value="a.value">{{ a.label }}</option>
            </select>
            <div v-if="newSite.auth_type === 'bearer'" class="flex gap-2">
              <input v-model="newAuth.token" placeholder="Token" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
            </div>
            <div v-if="newSite.auth_type === 'basic'" class="flex gap-2">
              <input v-model="newAuth.username" placeholder="Username" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
              <input v-model="newAuth.password" type="password" placeholder="Password" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
            </div>
            <div v-if="newSite.auth_type === 'header'" class="flex gap-2">
              <input v-model="newAuth.header" placeholder="Header name (e.g. X-Api-Key)" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
              <input v-model="newAuth.value" placeholder="Header value" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
            </div>
            <div v-if="newSite.auth_type === 'query'" class="flex gap-2">
              <input v-model="newAuth.param" placeholder="Param name (e.g. api_key)" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
              <input v-model="newAuth.value" placeholder="Param value" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
            </div>
          </div>
          <p v-if="formError" class="text-xs text-red-400">{{ formError }}</p>
          <div class="flex gap-2 pt-1">
            <button @click="addSite" class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-500 transition-colors">Add Site</button>
            <button @click="showAddSite = false; formError = ''" class="px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200 transition-colors">Cancel</button>
          </div>
        </div>
        <button v-else @click="showAddSite = true" class="flex items-center gap-1.5 px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200 bg-zinc-900 border border-zinc-800 rounded-lg hover:border-zinc-700 transition-colors">
          <Plus class="w-3.5 h-3.5" /> Add Site
        </button>
      </div>
    </template>

    <!-- Notifications Tab -->
    <template v-if="tab === 'notifications'">
      <div class="space-y-4">
        <!-- Channels -->
        <div>
          <h2 class="text-sm font-semibold text-zinc-300 mb-2">Channels</h2>
          <div class="space-y-2">
            <div v-for="ch in channels" :key="ch.id" class="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3 flex items-center justify-between">
              <div class="flex items-center gap-3">
                <span class="px-1.5 py-0.5 text-xs rounded border" :class="ch.type === 'telegram' ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' : ch.type === 'slack' ? 'bg-purple-500/10 text-purple-400 border-purple-500/20' : 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20'">{{ ch.type }}</span>
                <span class="text-zinc-200 text-sm">{{ ch.name || ch.type }}</span>
                <span v-if="!ch.enabled" class="text-xs text-zinc-600">disabled</span>
              </div>
              <div class="flex items-center gap-2">
                <button @click="testChannel(ch.id)" class="p-1 text-zinc-500 hover:text-emerald-400" title="Send test"><Send class="w-3.5 h-3.5" /></button>
                <button @click="removeChannel(ch.id)" class="p-1 text-zinc-500 hover:text-red-400"><Trash2 class="w-3.5 h-3.5" /></button>
              </div>
            </div>
          </div>

          <div v-if="showAddChannel" class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 mt-2 space-y-3 max-w-xl">
            <div class="flex gap-3">
              <select v-model="newChannel.type" class="bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 w-40">
                <option v-for="t in channelTypes" :key="t" :value="t">{{ t.charAt(0).toUpperCase() + t.slice(1) }}</option>
              </select>
              <input v-model="newChannel.name" placeholder="Channel name (optional)" class="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
            </div>
            <!-- Telegram fields -->
            <div v-if="newChannel.type === 'telegram'" class="space-y-2">
              <input v-model="channelConfig.token" placeholder="Bot token (from @BotFather)" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
              <input v-model="channelConfig.chat_id" placeholder="Chat ID" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
            </div>
            <!-- Slack / Teams fields -->
            <div v-else class="space-y-2">
              <input v-model="channelConfig.webhook_url" placeholder="Webhook URL" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500" />
            </div>
            <p v-if="channelError" class="text-xs text-red-400">{{ channelError }}</p>
            <div class="flex gap-2">
              <button @click="addChannel" class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-500">Add Channel</button>
              <button @click="showAddChannel = false; channelError = ''" class="px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200">Cancel</button>
            </div>
          </div>
          <button v-else @click="showAddChannel = true" class="flex items-center gap-1.5 px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200 mt-2">
            <Plus class="w-3.5 h-3.5" /> Add Channel
          </button>
        </div>

        <!-- Rules -->
        <div>
          <h2 class="text-sm font-semibold text-zinc-300 mb-2">Alert Rules</h2>
          <div class="space-y-2">
            <div v-for="r in rules" :key="r.id" class="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3 flex items-center justify-between">
              <div class="flex items-center gap-3 text-sm">
                <span class="px-1.5 py-0.5 text-xs rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">{{ r.type }}</span>
                <span class="text-zinc-400">→</span>
                <span class="text-zinc-300">{{ channelName(r.channel_id) }}</span>
                <span class="text-zinc-600 text-xs">cooldown: {{ r.cooldown_m }}m</span>
              </div>
              <button @click="removeRule(r.id)" class="p-1 text-zinc-500 hover:text-red-400"><Trash2 class="w-3.5 h-3.5" /></button>
            </div>
          </div>

          <div v-if="showAddRule" class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 mt-2 space-y-3 max-w-xl">
            <div class="space-y-3">
              <select v-model.number="newRule.channel_id" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200">
                <option :value="0" disabled>Select channel</option>
                <option v-for="ch in channels" :key="ch.id" :value="ch.id">{{ ch.name || ch.type }}</option>
              </select>
              <div>
                <div class="text-xs text-zinc-500 mb-2">Send alerts for:</div>
                <div class="flex flex-wrap gap-3">
                  <label v-for="t in ruleTypes" :key="t" class="flex items-center gap-1.5 text-xs text-zinc-300 cursor-pointer">
                    <input type="checkbox" v-model="newRule.types[t]" class="rounded border-zinc-600 bg-zinc-800 text-blue-500" />
                    {{ t.replace('_', ' ') }}
                  </label>
                </div>
              </div>
              <label class="flex items-center gap-2 text-xs text-zinc-400">Min interval between alerts: <input type="number" v-model.number="newRule.cooldown_m" min="1" class="w-16 bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-zinc-200 text-center" /> min</label>
            </div>
            <p v-if="ruleError" class="text-xs text-red-400">{{ ruleError }}</p>
            <div class="flex gap-2">
              <button @click="addRule" class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-500">Add Rules</button>
              <button @click="showAddRule = false; ruleError = ''" class="px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200">Cancel</button>
            </div>
          </div>
          <button v-else @click="showAddRule = true; newRule.types = initRuleTypes()" class="flex items-center gap-1.5 px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200 mt-2">
            <Plus class="w-3.5 h-3.5" /> Add Rules
          </button>
        </div>
      </div>
    </template>

    <!-- Schedules Tab -->
    <template v-if="tab === 'schedules'">
      <div class="space-y-3">
        <div v-for="hb in heartbeats" :key="hb.id" class="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3 flex items-center justify-between">
          <div class="flex items-center gap-3 text-sm">
            <Clock class="w-4 h-4 text-violet-400" />
            <span class="text-zinc-200 font-mono">{{ hb.command }}</span>
            <span v-if="hb.mode === 'at' && hb.at_time" class="text-zinc-500 text-xs">daily at {{ hb.at_time }}, grace {{ hb.grace_m }}m</span>
            <span v-else class="text-zinc-500 text-xs">every {{ hb.max_gap_m - hb.grace_m >= 60 ? Math.floor((hb.max_gap_m - hb.grace_m) / 60) + 'h' + ((hb.max_gap_m - hb.grace_m) % 60 ? ' ' + (hb.max_gap_m - hb.grace_m) % 60 + 'm' : '') : (hb.max_gap_m - hb.grace_m) + 'm' }}, grace {{ hb.grace_m }}m</span>
            <span v-if="hb.alerted" class="px-1.5 py-0.5 text-xs rounded bg-red-500/10 text-red-400 border border-red-500/20">Missed!</span>
            <span v-else-if="hb.last_seen_at" class="text-zinc-600 text-xs">last: {{ new Date(hb.last_seen_at * 1000).toLocaleTimeString() }}</span>
          </div>
          <button @click="removeHeartbeat(hb.id)" class="p-1 text-zinc-500 hover:text-red-400"><Trash2 class="w-3.5 h-3.5" /></button>
        </div>

        <div v-if="showAddHeartbeat" class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3 max-w-lg">
          <div class="space-y-3">
            <input v-model="newHeartbeat.command" @input="checkCommand" list="known-commands" placeholder="Command name (e.g. orders:cleanup)" class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 font-mono placeholder-zinc-500" />
            <datalist id="known-commands">
              <option v-for="cmd in knownCommands" :key="cmd" :value="cmd" />
            </datalist>
            <p v-if="commandUnknown" class="text-xs text-amber-400">This command hasn't been seen in traces yet. Make sure it's not in your ignore list and has run at least once.</p>
            <div class="flex gap-2">
              <button @click="newHeartbeat.mode = 'every'" class="px-3 py-1 text-xs rounded-md transition-colors" :class="newHeartbeat.mode === 'every' ? 'bg-zinc-700 text-zinc-100' : 'text-zinc-400 hover:text-zinc-200 bg-zinc-800'">Runs every</button>
              <button @click="newHeartbeat.mode = 'at'" class="px-3 py-1 text-xs rounded-md transition-colors" :class="newHeartbeat.mode === 'at' ? 'bg-zinc-700 text-zinc-100' : 'text-zinc-400 hover:text-zinc-200 bg-zinc-800'">Daily at</button>
            </div>
            <!-- Every mode -->
            <div v-if="newHeartbeat.mode === 'every'" class="flex items-center gap-2 text-xs text-zinc-400">
              Every
              <input type="number" v-model.number="newHeartbeat.interval_h" min="0" class="w-14 bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-zinc-200 text-center" /> h
              <input type="number" v-model.number="newHeartbeat.interval_m" min="0" class="w-14 bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-zinc-200 text-center" /> m
            </div>
            <!-- At mode -->
            <div v-else class="flex items-center gap-2 text-xs text-zinc-400">
              Daily at
              <input type="time" v-model="newHeartbeat.at_time" class="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-zinc-200" />
            </div>
            <div class="flex items-center gap-2 text-xs text-zinc-400">
              Grace period:
              <input type="number" v-model.number="newHeartbeat.grace_m" min="1" class="w-14 bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-zinc-200 text-center" /> min
            </div>
          </div>
          <p v-if="heartbeatError" class="text-xs text-red-400">{{ heartbeatError }}</p>
          <div class="flex gap-2">
            <button @click="addHeartbeat" class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-500">Add</button>
            <button @click="showAddHeartbeat = false; heartbeatError = ''" class="px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200">Cancel</button>
          </div>
        </div>
        <button v-else @click="showAddHeartbeat = true" class="flex items-center gap-1.5 px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200 bg-zinc-900 border border-zinc-800 rounded-lg">
          <Plus class="w-3.5 h-3.5" /> Add Schedule Watch
        </button>
      </div>
    </template>
  </div>
</template>
