<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type Email } from '@/api'

const props = defineProps<{ id: string }>()
const email = ref<Email | null>(null)
const showHtml = ref(true)

onMounted(async () => {
  email.value = await api.mailDetail(props.id)
})
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
        <h3 class="text-sm font-semibold mb-2">Attachments</h3>
        <div v-for="a in email.attachments" :key="a.name" class="text-sm text-zinc-400">
          {{ a.name }} ({{ a.content_type }}, {{ (a.size / 1024).toFixed(1) }}KB)
        </div>
      </div>
    </template>
  </div>
</template>
