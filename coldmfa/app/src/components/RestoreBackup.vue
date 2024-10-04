<script setup lang="ts">
import type { AxiosInstance } from 'axios'
import { inject, onMounted, ref, useTemplateRef } from 'vue'

const client = inject<AxiosInstance>('client') as AxiosInstance

const password = ref('')
const fileContent = ref('')
const errMsg = ref('')

const input = useTemplateRef('passwordInput')

onMounted(() => {
  input.value?.focus()
})

const loadInputFile = (ev: Event) => {
  const file = (ev.target as HTMLInputElement | null)?.files?.[0]
  if (!file) return

  const reader = new FileReader()

  reader.onload = (e) => {
    fileContent.value = e.target?.result as string
  }
  reader.readAsText(file)
}

const downloadBackup = async () => {
  try {
    await client.put(
      'api/backups',
      {
        backupContent: Array.from(new TextEncoder().encode(fileContent.value)),
        password: password.value
      },
      { validateStatus: (status) => status === 200 }
    )

    window.location.reload()
  } catch (e) {
    console.error(e)
  }
}
</script>

<template>
  <p class="text-xl bold">Enter a password to decrypt the backup</p>
  <form class="flex flex-col py-2" @submit.prevent.stop="downloadBackup">
    <input
      type="text"
      placeholder="Password"
      class="input input-bordered input-accent w-full max-w-s"
      autocomplete="off"
      ref="passwordInput"
      v-model="password"
    />
    <div class="flex flex-row mx-auto mt-5">
      <input type="file" class="file-input w-full max-w-xs" @change="loadInputFile" />
    </div>
    <p v-if="errMsg" class="text-red-500 py-2">Error restoring from backup: {{ errMsg }}</p>

    <div class="flex flex-row justify-end my-2 mx-1">
      <button class="btn btn-primary rounded p-2 mt-2" type="submit">Restore</button>
    </div>
  </form>
</template>

<style scoped></style>
