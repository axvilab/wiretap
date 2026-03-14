<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type Email } from '@/api'

const props = defineProps<{ id: string }>()
const email = ref<Email | null>(null)
const showHtml = ref(true)

onMounted(async () => {
  email.value = await api.mailDetail(props.id)
})

async function downloadFile(url: string, filename: string) {
  const res = await fetch(url)
  const blob = await res.blob()
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = filename
  a.click()
  URL.revokeObjectURL(a.href)
}
</script>

<template>
  <div class="p-6 space-y-4">
    <button @click="$router.back()" class="text-zinc-400 hover:text-zinc-100 text-sm">&larr; Back to Mail</button>

    <template v-if="email">
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-2">
        <h1 class="text-xl font-bold">{{ email.subject }}</h1>
        <div class="text-sm text-zinc-400 space-y-1">
          <div>From: <span class="text-zinc-200">{{ email.from }}</span></div>
          <div>To: <span class="text-zinc-200">{{ email.to.join(', ') }}</span></div>
          <div v-if="email.cc?.length">Cc: <span class="text-zinc-200">{{ email.cc.join(', ') }}</span></div>
        </div>
      </div>

      <!-- Toggle -->
      <div v-if="email.html && email.text" class="flex gap-2">
        <button
          @click="showHtml = true"
          class="px-3 py-1 text-sm rounded"
          :class="showHtml ? 'bg-zinc-700 text-zinc-100' : 'bg-zinc-800 text-zinc-400'"
        >HTML</button>
        <button
          @click="showHtml = false"
          class="px-3 py-1 text-sm rounded"
          :class="!showHtml ? 'bg-zinc-700 text-zinc-100' : 'bg-zinc-800 text-zinc-400'"
        >Text</button>
      </div>

      <!-- Content -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <iframe
          v-if="showHtml && email.html"
          :srcdoc="email.html"
          class="w-full min-h-[400px] bg-white"
          sandbox="allow-same-origin"
        />
        <pre v-else class="p-4 text-sm text-zinc-300 whitespace-pre-wrap">{{ email.text || '(no text content)' }}</pre>
      </div>

      <!-- Attachments -->
      <div v-if="email.attachments?.length" class="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <h3 class="text-sm font-semibold mb-2">Attachments ({{ email.attachments.length }})</h3>
        <div v-for="a in email.attachments" :key="a.index" class="flex items-center justify-between py-2 border-b border-zinc-800/50 last:border-0">
          <div class="text-sm">
            <span class="text-zinc-200">{{ a.name }}</span>
            <span class="text-zinc-500 ml-2">{{ a.content_type }}</span>
            <span class="text-zinc-600 ml-2">{{ (a.size / 1024).toFixed(1) }}KB</span>
          </div>
          <button
            @click="downloadFile(`/api/mail/${email.id}/attachment?index=${a.index}`, a.name)"
            class="px-3 py-1 text-xs bg-zinc-800 hover:bg-zinc-700 rounded text-zinc-300 transition-colors cursor-pointer"
          >Download</button>
        </div>
      </div>

      <!-- Raw email download -->
      <div class="flex gap-2">
        <button
          @click="downloadFile(`/api/mail/${email.id}/raw`, `${email.subject.replace(/[^a-zA-Z0-9]/g, '_')}.eml`)"
          class="px-3 py-1.5 text-xs bg-zinc-800 hover:bg-zinc-700 rounded text-zinc-400 transition-colors cursor-pointer"
        >Download .eml</button>
      </div>
    </template>
  </div>
</template>
