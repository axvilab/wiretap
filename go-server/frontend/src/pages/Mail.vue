<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Email } from '@/api'
import Pagination from '@/components/Pagination.vue'
import NewItemsBanner from '@/components/NewItemsBanner.vue'

const emails = ref<Email[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(25)

async function load() {
  const data = await api.mail(`page=${page.value}&per_page=${perPage.value}`)
  emails.value = data.emails
  total.value = data.total
}

onMounted(load)
watch([page, perPage], load)
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

    <NewItemsBanner :mail-only="true" @refresh="load" />

    <div v-if="!emails.length" class="text-zinc-500 py-8 text-center">No emails intercepted yet</div>

    <div v-if="emails.length" class="bg-zinc-900 border border-zinc-800 rounded-lg">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-zinc-500 text-xs border-b border-zinc-800">
            <th class="px-4 py-2 text-left">Time</th>
            <th class="px-4 py-2 text-left">Subject</th>
            <th class="px-4 py-2 text-left">From</th>
            <th class="px-4 py-2 text-left">To</th>
            <th class="px-4 py-2 text-right">Attachments</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="e in emails"
            :key="e.id"
            class="border-b border-zinc-800/50 hover:bg-zinc-800/30 cursor-pointer transition-colors"
            @click="$router.push(`/mail/${e.id}`)"
          >
            <td class="px-4 py-2 text-zinc-500 font-mono text-xs">{{ timeAgo(e.created_at) }}</td>
            <td class="px-4 py-2 text-sm font-medium">{{ e.subject }}</td>
            <td class="px-4 py-2 text-xs text-zinc-400">{{ e.from }}</td>
            <td class="px-4 py-2 text-xs text-zinc-400">{{ e.to?.join(', ') }}</td>
            <td class="px-4 py-2 text-right">
              <span v-if="e.attachments?.length" class="inline-flex px-1.5 py-0.5 text-xs rounded border bg-emerald-500/10 text-emerald-400 border-emerald-500/20">
                {{ e.attachments.length }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-model:page="page" v-model:per-page="perPage" :total="total" />
  </div>
</template>
