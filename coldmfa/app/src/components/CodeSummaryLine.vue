<script setup lang="ts">
import type { ApiError, CodeSummary, PasscodeResponse } from '@/types'
import type { AxiosInstance, AxiosResponse } from 'axios'
import { inject, ref } from 'vue'
import CodeTicker from '@/components/CodeTicker.vue'

const props = defineProps<{
  groupId: string
  code: CodeSummary
}>()

defineEmits<{
  showExport: [codeId: string]
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance
const fetchedCode = ref<PasscodeResponse | undefined>()

const getCode = async () => {
  try {
    const codesResponse: AxiosResponse<PasscodeResponse> = await client.get(
      `api/groups/${props.groupId}/codes/${props.code.codeId}`
    )

    if (codesResponse.status === 200) {
      fetchedCode.value = codesResponse.data
    } else {
      console.error(codesResponse.data)
    }
  } catch (e) {
    const error = e as ApiError
    console.error(error)
  }
}

const clearExpiredCode = (serverTime: number) => {
  if (fetchedCode.value && fetchedCode.value.serverTime == serverTime) {
    fetchedCode.value = undefined
  }
}
</script>

<template>
  <div class="flex flex-row w-full">
    <p class="w-1/3">{{ props.code.preferredName ?? props.code.name }}</p>
    <div class="flex w-1/3 justify-center">
      <template v-if="fetchedCode">
        <CodeTicker :passcode-response="fetchedCode" @expired="clearExpiredCode"></CodeTicker>
      </template>
    </div>
    <div class="flex w-1/3 justify-end">
      <div class="join">
        <button class="btn btn-secondary join-item" @click="$emit('showExport', props.code.codeId)">
          Export
        </button>
        <button class="btn btn-primary join-item" @click="getCode">Get a code</button>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
