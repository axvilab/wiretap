<script setup lang="ts">
const props = defineProps<{ method: string }>()

const colors: Record<string, string> = {
  GET: 'text-sky-400',
  POST: 'text-emerald-400',
  PUT: 'text-amber-400',
  PATCH: 'text-orange-400',
  DELETE: 'text-red-400',
}
const color = colors[props.method] || 'text-zinc-400'
</script>

<template>
  <span class="font-mono text-xs font-bold" :class="color">{{ method }}</span>
</template>
