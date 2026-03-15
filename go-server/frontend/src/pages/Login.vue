<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ setup?: boolean }>()
const emit = defineEmits<{ success: [] }>()

const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''

  if (props.setup && password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }

  if (password.value.length < 6) {
    error.value = 'Password must be at least 6 characters'
    return
  }

  loading.value = true
  try {
    const endpoint = props.setup ? '/api/auth/setup' : '/api/auth/login'
    const res = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: password.value }),
    })
    const data = await res.json()

    if (!res.ok) {
      error.value = data.error || 'Authentication failed'
      return
    }

    emit('success')
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-zinc-950 flex items-center justify-center">
    <div class="w-full max-w-sm">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-zinc-100">Wiretap</h1>
        <p class="text-zinc-500 mt-2">{{ setup ? 'Create your password' : 'Sign in to continue' }}</p>
      </div>

      <form @submit.prevent="submit" class="bg-zinc-900 border border-zinc-800 rounded-lg p-6 space-y-4">
        <div>
          <label class="text-xs text-zinc-400 mb-1 block">Password</label>
          <input
            v-model="password"
            type="password"
            autofocus
            class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
            :placeholder="setup ? 'Create a password (min 6 chars)' : 'Enter your password'"
          />
        </div>

        <div v-if="setup">
          <label class="text-xs text-zinc-400 mb-1 block">Confirm Password</label>
          <input
            v-model="confirmPassword"
            type="password"
            class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
            placeholder="Confirm your password"
          />
        </div>

        <div v-if="error" class="bg-red-500/10 border border-red-500/20 rounded px-3 py-2 text-sm text-red-400">
          {{ error }}
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 rounded font-medium text-sm transition-colors"
        >
          {{ loading ? 'Please wait...' : (setup ? 'Create Password' : 'Sign In') }}
        </button>
      </form>
    </div>
  </div>
</template>
