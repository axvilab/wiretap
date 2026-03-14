<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type Email } from '@/api'

const emails = ref<Email[]>([])
const total = ref(0)

onMounted(async () => {
  const data = await api.mail()
  emails.value = data.emails
  total.value = data.total
})

function timeAgo(ts: number): string {
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}
</script>

<template>
  <div class="p-6 space-y-4">
    <h1 class="text-2xl font-bold">Mail ({{ total }})</h1>

    <div v-if="!emails.length" class="text-zinc-500 py-8 text-center">No emails intercepted yet</div>

    <div
      v-for="e in emails"
      :key="e.id"
      class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 cursor-pointer hover:border-zinc-700 transition-colors"
      @click="$router.push(`/mail/${e.id}`)"
    >
      <div class="flex items-center justify-between mb-1">
        <span class="font-semibold text-sm">{{ e.subject }}</span>
        <span class="text-zinc-500 text-xs">{{ timeAgo(e.created_at) }}</span>
      </div>
      <div class="text-zinc-400 text-xs">
        <span>From: {{ e.from }}</span>
        <span class="mx-2">&rarr;</span>
        <span>To: {{ e.to.join(', ') }}</span>
      </div>
      <div v-if="e.attachments?.length" class="text-zinc-500 text-xs mt-1">
        {{ e.attachments.length }} attachment(s)
      </div>
    </div>
  </div>
</template>
