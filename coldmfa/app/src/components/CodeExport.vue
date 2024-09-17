<script setup lang="ts">
import type { AxiosInstance } from 'axios'
import { inject, ref, watch } from 'vue'
import type { CodeSummary } from '@/types'

const props = defineProps<{
  groupId: string
  code: CodeSummary
}>()

defineEmits<{
  close: [x: void]
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance

const imgSrc = ref('')
const autoHidden = ref(false)

const fetchQR = async (): Promise<ArrayBuffer | null> => {
  try {
    const response = await client.get(`api/groups/${props.groupId}/codes/${props.code.codeId}/qr`, {
      responseType: 'arraybuffer'
    })
    if (response.status === 200) {
      return response.data
    } else {
      console.error(response)
    }
  } catch (e) {
    console.error(e)
  }

  return null
}

const revealImage = () => {
  autoHidden.value = false
  setTimeout(() => {
    autoHidden.value = true
  }, 5_000)
}

watch(
  () => props.groupId,
  () => {
    imgSrc.value = ''
  }
)

watch(
  () => props.code,
  () => {
    fetchQR().then((bytes) => {
      if (bytes) {
        //@ts-ignore
        const imgBase64 = btoa(String.fromCharCode.apply(null, new Uint8Array(bytes)))
        imgSrc.value = `data:image/jpg;base64,${imgBase64}`
        revealImage()
      }
    })
  },
  { immediate: true }
)
</script>

<template>
  <div class="flex flex-col">
    <p class="mx-auto mb-2 text-accent text-2xl">{{ code.preferredName ?? code.name }}</p>
    <p class="mx-auto cursor-pointer" v-if="autoHidden" @click="revealImage">Hidden...</p>
    <img v-else alt="QR code" v-bind:src="imgSrc" />
    <button class="btn btn-accent my-5" @click="$emit('close')">Close</button>
  </div>
</template>

<style scoped></style>
