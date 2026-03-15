<script setup lang="ts">
const props = defineProps<{ method: string }>()

const styles: Record<string, string> = {
  GET: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
  POST: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
  PUT: 'bg-sky-500/10 text-sky-400 border-sky-500/20',
  PATCH: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
  DELETE: 'bg-red-500/10 text-red-400 border-red-500/20',
}
const style = styles[props.method] || 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20'
</script>

<template>
  <span class="inline-flex px-2 py-0.5 text-xs font-mono font-bold rounded border" :class="style">{{ method }}</span>
</template>
