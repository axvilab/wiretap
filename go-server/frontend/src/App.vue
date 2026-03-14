<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { LayoutDashboard, Radio, Bug, Briefcase, Mail, Activity } from 'lucide-vue-next'

const nav = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/requests', label: 'Requests', icon: Activity },
  { to: '/exceptions', label: 'Exceptions', icon: Bug },
  { to: '/jobs', label: 'Jobs', icon: Briefcase },
  { to: '/mail', label: 'Mail', icon: Mail },
]
</script>

<template>
  <div class="flex h-screen">
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
      <div class="p-4 border-t border-zinc-800 text-xs text-zinc-600">
        Wiretap v0.1
      </div>
    </nav>

    <!-- Main content -->
    <main class="flex-1 overflow-auto">
      <RouterView />
    </main>
  </div>
</template>
