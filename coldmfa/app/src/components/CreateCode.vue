<script setup lang="ts">
import { inject, onMounted, ref, useTemplateRef } from 'vue'
import type { AxiosError, AxiosInstance, AxiosResponse } from 'axios'
import type { ApiError, CodeSummary } from '@/types'
import { useGroupsStore } from '@/stores/groups'

const props = defineProps<{
  groupId: string
}>()

const emit = defineEmits<{
  created: [code: CodeSummary]
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance
const groupsStore = useGroupsStore()

const input = useTemplateRef('codeNameInput')

const original = ref('')
const errMsg = ref('')

onMounted(() => {
  input.value?.focus()
})

const storeCode = async (event: Event) => {
  event.preventDefault()
  event.stopPropagation()

  try {
    const groupId = props.groupId
    const response: AxiosResponse<CodeSummary | ApiError> = await client.post(
      `api/groups/${groupId}/codes`,
      {
        original: original.value
      }
    )

    if (response.status === 201) {
      original.value = ''
      groupsStore.addCodeToGroup(groupId, response.data as CodeSummary)
      emit('created', response.data as CodeSummary)
    } else {
      errMsg.value = (response.data as ApiError).error
    }
  } catch (error) {
    const err = error as AxiosError
    if (
      err?.response &&
      err?.response?.data &&
      typeof err.response.data === 'object' &&
      'error' in err.response.data
    ) {
      errMsg.value = (err.response.data as ApiError).error
    } else {
      errMsg.value = 'Unknown error'
      console.error(err)
    }
  }
}
</script>

<template>
  <p class="text-xl bold">Create a new code</p>
  <form class="flex flex-col py-2" @submit="storeCode">
    <input
      type="text"
      placeholder="URL for the One Time Password"
      name="original"
      class="input input-bordered input-accent w-full max-w-s"
      ref="codeNameInput"
      v-model="original"
    />
    <p v-if="errMsg" class="text-red-500 py-2">Error creating your group: {{ errMsg }}</p>

    <div class="flex flex-row justify-end my-2 mx-1">
      <button class="btn btn-primary rounded p-2 mt-2" type="submit">Store code</button>
    </div>
  </form>
</template>

<style scoped></style>
