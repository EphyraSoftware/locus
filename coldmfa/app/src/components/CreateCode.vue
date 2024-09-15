<script setup lang="ts">
import { ref } from 'vue'
import type { AxiosInstance, AxiosResponse } from 'axios'
import type { ApiError, CodeSummary } from '@/types'

const props = defineProps<{
  client: AxiosInstance
  groupId: string
}>()

const original = ref('')

const storeCode = async (event: Event) => {
  event.preventDefault()
  event.stopPropagation()

  try {
    const response: AxiosResponse<CodeSummary | ApiError> = await props.client.post(
      `api/groups/${props.groupId}/codes`,
      {
        original: original.value
      }
    )

    if (response.status === 201) {
      console.log('Code stored successfully', response.data as CodeSummary)
    } else {
      console.error('Failed to store code', response.data as ApiError)
    }
  } catch (error) {
    console.error(error)
  }
}
</script>

<template>
  <div>
    <p class="text-xl bold">Create a new code</p>
    <form class="flex flex-col" @submit="storeCode">
      <label for="original">URL for the One Time Password</label>
      <input type="text" id="original" name="original" class="rounded" v-model="original" />

      <div class="flex flex-row justify-end my-5 mx-1">
        <button class="bg-teal-400 rounded p-2 mt-2" type="submit">Store code</button>
      </div>
    </form>
  </div>
</template>

<style scoped></style>
