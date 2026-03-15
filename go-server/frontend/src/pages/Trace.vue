<script setup lang="ts">
import { ref, watch } from 'vue'
import { api, type TraceNode } from '@/api'
import StatusBadge from '@/components/StatusBadge.vue'
import MethodBadge from '@/components/MethodBadge.vue'
import TraceTree from '@/components/TraceTree.vue'

const props = defineProps<{ rid: string }>()
const tree = ref<TraceNode | null>(null)
const error = ref('')

async function loadTrace(rid: string) {
  tree.value = null
  error.value = ''
  try {
    tree.value = await api.trace(rid)
  } catch (e: any) {
    error.value = e.message
  }
}

watch(() => props.rid, (rid) => loadTrace(rid), { immediate: true })

function parseWarnings(json: string) {
  try { return JSON.parse(json) } catch { return [] }
}

function parseMeta(json: string) {
  try { return JSON.parse(json) } catch { return {} }
}
</script>

<template>
  <div class="p-6 space-y-6">
    <div v-if="error" class="text-red-400">{{ error }}</div>

    <template v-if="tree">
      <!-- Header -->
      <div class="flex items-center gap-3">
        <button @click="$router.back()" class="text-zinc-400 hover:text-zinc-100 text-2xl leading-none">&larr;</button>
        <MethodBadge v-if="tree.trace.method" :method="tree.trace.method" />
        <span class="font-mono text-lg">{{ tree.trace.uri || parseMeta(tree.trace.meta_json).job_class || tree.trace.type }}</span>
        <StatusBadge v-if="tree.trace.status" :status="tree.trace.status" />
        <span class="text-zinc-500 text-sm">{{ tree.trace.duration_ms.toFixed(1) }}ms</span>
      </div>

      <!-- Warnings -->
      <div v-if="parseWarnings(tree.trace.warnings_json).length" class="space-y-1">
        <div
          v-for="(w, i) in parseWarnings(tree.trace.warnings_json)"
          :key="i"
          class="bg-amber-500/10 border border-amber-500/20 rounded px-3 py-2 text-sm"
        >
          <span class="text-amber-400 font-semibold">{{ w.type }}</span>
          <span class="text-zinc-300 ml-2">{{ w.message }}</span>
          <span v-if="w.suggestion" class="text-zinc-500 ml-2">— {{ w.suggestion }}</span>
        </div>
      </div>

      <!-- Trace tree -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <TraceTree :node="tree" :depth="0" />
      </div>
    </template>
  </div>
</template>
