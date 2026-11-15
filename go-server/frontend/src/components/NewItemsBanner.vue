<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useWebSocket, type WSEvent } from '@/hooks/useWebSocket'

const props = defineProps<{
  /** Trace types to count, e.g. ['request'] or ['job'] or empty for all */
  types?: string[]
  /** Only count traces with exceptions */
  exceptionsOnly?: boolean
  /** Count only traces containing mail.sent events */
  mailOnly?: boolean
  /** Count only traces containing broadcast events */
  broadcastOnly?: boolean
  /** Count only traces containing log events */
  logsOnly?: boolean
}>()

const emit = defineEmits<{
  refresh: []
}>()

const count = ref(0)

useWebSocket((event: WSEvent) => {
  if (event.type !== 'trace.new' || !event.data) return

  // Filter by type if specified
  if (props.types && props.types.length > 0) {
    if (!props.types.includes(event.data.type)) return
  }


  // Filter by exception if specified
  if (props.exceptionsOnly && !event.data.has_exception) return

  // Filter for mail — check if trace has mail events
  if (props.mailOnly) {
    try {
      const events = JSON.parse(event.data.events_json || '[]')
      if (!events.some((e: any) => e.type === 'mail.sent')) return
    } catch { return }
  }

  // Filter for broadcasts
  if (props.broadcastOnly) {
    try {
      const events = JSON.parse(event.data.events_json || '[]')
      if (!events.some((e: any) => e.type === 'broadcast')) return
    } catch { return }
  }

  // Filter for logs
  if (props.logsOnly) {
    try {
      const events = JSON.parse(event.data.events_json || '[]')
      if (!events.some((e: any) => e.type === 'log')) return
    } catch { return }
  }

  count.value++
})

function handleClick() {
  count.value = 0
  emit('refresh')
}
</script>

<template>
  <div
    v-if="count > 0"
    @click="handleClick"
    class="flex items-center justify-center gap-2 px-4 py-2 bg-indigo-500/10 border border-indigo-500/20 rounded-lg cursor-pointer hover:bg-indigo-500/20 transition-colors"
  >
    <span class="text-sm text-indigo-400">
      {{ count }} new {{ count === 1 ? 'item' : 'items' }} available
    </span>
    <span class="text-xs text-indigo-500">Click to refresh</span>
  </div>
</template>