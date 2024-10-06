<script setup lang="ts">
import type { AxiosInstance } from 'axios'
import { inject, nextTick, onMounted, ref, useTemplateRef } from 'vue'

const emit = defineEmits<{
  completed: [x: void]
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance

const password = ref('')
const errMsg = ref('')

const downloadContent = ref('')

const input = useTemplateRef('passwordInput')
const downloadAnchor = useTemplateRef('download')

onMounted(() => {
  input.value?.focus()
})

const downloadBackup = async () => {
  try {
    const backup = await client.post(
      'api/backups',
      { password: password.value },
      { validateStatus: (status) => status === 200 }
    )

    downloadContent.value = backup.data
    await nextTick()
    downloadAnchor.value?.click()
    emit('completed')
  } catch (e) {
    console.error(e)
  }
}
</script>

<template>
  <p class="text-xl bold">Enter a password to encrypt the backup with</p>
  <form class="flex flex-col py-2" @submit.prevent.stop="downloadBackup">
    <input
      type="password"
      placeholder="Password"
      class="input input-bordered input-accent w-full max-w-s"
      autocomplete="off"
      ref="passwordInput"
      v-model="password"
    />
    <p v-if="errMsg" class="text-red-500 py-2">Error preparing backup: {{ errMsg }}</p>

    <div class="flex flex-row justify-end my-2 mx-1">
      <button class="btn btn-primary rounded p-2 mt-2" type="submit">Download</button>
    </div>
  </form>

  <a
    ref="download"
    class="hidden"
    download="cold-mfa-backup.age.txt"
    :href="'data:text/plain;charset=utf-8,' + encodeURIComponent(downloadContent)"
  ></a>
</template>

<style scoped></style>
