<script setup lang="ts">
import type { Trace } from '@/api'

defineProps<{ trace: Trace }>()
</script>

<template>
  <div class="flex gap-1">
    <span v-if="trace.has_exception" title="Exception" class="text-red-400">✕</span>
    <span v-if="trace.has_n_plus_one" title="N+1 Query" class="text-amber-400">N+1</span>
    <span v-if="trace.has_slow_query" title="Slow Query" class="text-orange-400">⏱</span>
    <span v-if="trace.has_duplicate_query" title="Duplicate Query" class="text-purple-400">⊘</span>
  </div>
</template>
