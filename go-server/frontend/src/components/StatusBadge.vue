<script setup lang="ts">
const props = defineProps<{ status: number }>()

const color = props.status >= 500
  ? 'bg-red-500/10 text-red-400 border-red-500/20'
  : props.status >= 400
    ? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
    : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
</script>

<template>
  <span class="inline-flex px-2 py-0.5 text-xs font-mono rounded border" :class="color">
    {{ status }}
  </span>
</template>
