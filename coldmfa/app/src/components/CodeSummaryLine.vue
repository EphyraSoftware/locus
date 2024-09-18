<script setup lang="ts">
import type { ApiError, CodeSummary, PasscodeResponse } from '@/types'
import type { AxiosInstance, AxiosResponse } from 'axios'
import { inject, ref, useTemplateRef } from 'vue'
import CodeTicker from '@/components/CodeTicker.vue'
import { useGroupsStore } from '@/stores/groups'

const props = defineProps<{
  groupId: string
  codeId: string
}>()

defineEmits<{
  showExport: [codeId: string]
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance
const groupsStore = useGroupsStore()
const fetchedCode = ref<PasscodeResponse | undefined>()

const codeName = useTemplateRef<HTMLParagraphElement>('codeName')

const code = groupsStore.codeById(props.groupId, props.codeId)

const getCode = async () => {
  if (!code) {
    return
  }

  try {
    const codesResponse: AxiosResponse<PasscodeResponse> = await client.get(
      `api/groups/${props.groupId}/codes/${code.codeId}`
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

const renameCode = async () => {
  if (!code) {
    return
  }

  try {
    let newPreferredName = codeName.value?.textContent
    if (newPreferredName === null) {
      return
    }

    const response: AxiosResponse<CodeSummary | ApiError> = await client.put(
      `api/groups/${props.groupId}/codes/${code.codeId}`,
      {
        preferredName: newPreferredName
      }
    )

    if (response.status === 204) {
      if (newPreferredName === '') {
        code.preferredName = undefined
      } else {
        code.preferredName = newPreferredName
      }
      groupsStore.replaceCodeInGroup(props.groupId, code)
    } else {
      console.error(response)
    }
  } catch (e) {
    const error = e as ApiError
    console.error(error)
  }
}
</script>

<template>
  <div class="flex flex-row w-full">
    <div class="w-1/3">
      <p class="inline-block" contenteditable="true" ref="codeName" @focusout="renameCode">
        {{ code?.preferredName ?? code?.name }}
      </p>
    </div>
    <div class="flex w-1/3 justify-center">
      <template v-if="fetchedCode">
        <CodeTicker :passcode-response="fetchedCode" @expired="clearExpiredCode"></CodeTicker>
      </template>
    </div>
    <div class="flex w-1/3 justify-end">
      <div class="join">
        <button
          class="btn btn-secondary join-item"
          @click="
            () => {
              if (code) {
                $emit('showExport', code.codeId)
              }
            }
          "
        >
          Export
        </button>
        <button class="btn btn-primary join-item" @click="getCode">Get a code</button>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
