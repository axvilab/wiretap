<script setup lang="ts">
const PAGE_SIZES = [25, 50, 100] as const

const page = defineModel<number>('page', { required: true })
const perPage = defineModel<number>('perPage', { required: true })
const props = defineProps<{ total: number }>()

const totalPages = () => Math.max(1, Math.ceil(props.total / perPage.value))

function setPerPage(value: number) {
  perPage.value = value
  page.value = 1
}
</script>

<template>
  <div class="flex items-center justify-center gap-2">
    <button
      :disabled="page <= 1"
      @click="page--"
      class="px-3 py-1 text-sm bg-zinc-800 rounded hover:bg-zinc-700 disabled:opacity-40"
    >Prev</button>
    <span class="text-sm text-zinc-400">{{ page }} / {{ totalPages() }}</span>
    <button
      :disabled="page >= totalPages()"
      @click="page++"
      class="px-3 py-1 text-sm bg-zinc-800 rounded hover:bg-zinc-700 disabled:opacity-40"
    >Next</button>
    <select
      :value="perPage"
      @change="setPerPage(Number(($event.target as HTMLSelectElement).value))"
      class="ml-4 bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-sm"
    >
      <option v-for="size in PAGE_SIZES" :key="size" :value="size">{{ size }}</option>
    </select>
  </div>
</template>
