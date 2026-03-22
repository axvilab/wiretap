<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { LayoutDashboard, Radio, Bug, Briefcase, Mail, Activity, Podcast, Settings, Terminal, Clock, LogOut, FileText, Bell, Gauge, List } from 'lucide-vue-next'
import Login from '@/pages/Login.vue'

const nav = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/activity', label: 'Activity', icon: List },
  { to: '/performance', label: 'Performance', icon: Gauge },
  { to: '/requests', label: 'Requests', icon: Activity },
  { to: '/commands', label: 'Commands', icon: Terminal },
  { to: '/schedule', label: 'Schedule', icon: Clock },
  { to: '/jobs', label: 'Jobs', icon: Briefcase },
  { to: '/exceptions', label: 'Exceptions', icon: Bug },
  { to: '/mail', label: 'Mail', icon: Mail },
  { to: '/broadcasts', label: 'Broadcasts', icon: Podcast },
  { to: '/notifications', label: 'Notifications', icon: Bell },
  { to: '/logs', label: 'Logs', icon: FileText },
  { to: '/settings', label: 'Settings', icon: Settings },
]

const authState = ref<'loading' | 'login' | 'ready'>('loading')

async function checkAuth() {
  try {
    const res = await fetch('/api/auth/status')
    const data = await res.json()

    if (!data.auth_enabled || data.authenticated) {
      authState.value = 'ready'
    } else {
      authState.value = 'login'
    }
  } catch {
    authState.value = 'ready'
  }
}

async function logout() {
  await fetch('/api/auth/logout', { method: 'POST' })
  authState.value = 'login'
}

onMounted(checkAuth)
</script>

<template>
  <!-- Loading -->
  <div v-if="authState === 'loading'" class="min-h-screen bg-zinc-950 flex items-center justify-center">
    <div class="text-zinc-500">Loading...</div>
  </div>

  <!-- Login -->
  <Login v-else-if="authState === 'login'" @success="authState = 'ready'" />

  <!-- App -->
  <div v-else class="flex h-screen">
    <!-- Sidebar -->
    <nav class="w-56 shrink-0 border-r border-zinc-800 bg-zinc-900 flex flex-col">
      <div class="p-4 border-b border-zinc-800">
        <RouterLink to="/" class="flex items-center gap-2">
          <Radio class="w-6 h-6 text-emerald-400" />
          <span class="text-lg font-bold tracking-tight">Wiretap</span>
        </RouterLink>
      </div>
      <div class="flex-1 py-2">
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 px-4 py-2.5 text-sm text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/50 transition-colors"
          active-class="!text-emerald-400 bg-zinc-800/50"
        >
          <component :is="item.icon" class="w-4 h-4" />
          {{ item.label }}
        </RouterLink>
      </div>
      <div class="p-4 border-t border-zinc-800 flex items-center justify-between">
        <span class="text-xs text-zinc-600">Wiretap v0.1</span>
        <button @click="logout" class="text-zinc-600 hover:text-zinc-400 transition-colors" title="Sign out">
          <LogOut class="w-4 h-4" />
        </button>
      </div>
    </nav>

    <!-- Main content -->
    <main class="flex-1 overflow-auto">
      <RouterView />
    </main>
  </div>
</template>
