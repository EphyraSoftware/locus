<script setup lang="ts">
import type { ApiError, CodeSummary, PasscodeResponse } from '@/types'
import type { AxiosInstance, AxiosResponse } from 'axios'
import { computed, inject, ref, useTemplateRef } from 'vue'
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

const code = computed(() => groupsStore.codeById(props.groupId, props.codeId))

const deleteCounter = ref(5)

const getCode = async () => {
  try {
    const codesResponse: AxiosResponse<PasscodeResponse> = await client.get(
      `api/groups/${props.groupId}/codes/${props.codeId}`
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
  const currentCode = code.value
  if (!currentCode) {
    return
  }

  try {
    let newPreferredName = codeName.value?.textContent
    if (
      newPreferredName === null ||
      newPreferredName === currentCode?.name ||
      newPreferredName === currentCode?.preferredName
    ) {
      return
    }

    const response: AxiosResponse<CodeSummary | ApiError> = await client.put(
      `api/groups/${props.groupId}/codes/${currentCode.codeId}`,
      {
        preferredName: newPreferredName
      }
    )

    if (response.status === 204) {
      if (newPreferredName === '') {
        currentCode.preferredName = undefined
      } else {
        currentCode.preferredName = newPreferredName
      }
      groupsStore.replaceCodeInGroup(props.groupId, currentCode)
    } else {
      console.error(response)
    }
  } catch (e) {
    const error = e as ApiError
    console.error(error)
  }
}

const tryDelete = async () => {
  const currentCode = code.value
  if (!currentCode) {
    return
  }

  if (deleteCounter.value == 5) {
    // After the first click, you have 5 seconds to confirm
    setTimeout(() => {
      deleteCounter.value = 5
    }, 5000)
  }

  if (deleteCounter.value == 0) {
    try {
      const response = await client.delete(`api/groups/${props.groupId}/codes/${props.codeId}`)
      if (response.status === 204) {
        currentCode.deleted = true
        currentCode.deletedAt = new Date().valueOf()
        groupsStore.replaceCodeInGroup(props.groupId, currentCode)
      }
    } catch (e) {
      console.error(e)
    }
  } else {
    deleteCounter.value--
  }
}
</script>

<template>
  <div class="flex flex-row w-full">
    <div class="w-1/3">
      <p
        class="inline-block"
        contenteditable="true"
        spellcheck="false"
        ref="codeName"
        @focusout="renameCode"
      >
        {{ code?.preferredName ?? code?.name }}
      </p>
    </div>
    <div class="flex w-1/3 justify-center">
      <template v-if="fetchedCode">
        <CodeTicker :passcode-response="fetchedCode" @expired="clearExpiredCode"></CodeTicker>
      </template>
      <p v-else-if="code?.deleted">This code has been deleted</p>
    </div>
    <div class="flex w-1/3 justify-end">
      <div class="join">
        <button class="btn btn-error join-item" @click="tryDelete" :disabled="code?.deleted">
          {{ deleteCounter == 5 ? 'Delete' : `Confirm? (${deleteCounter})` }}
        </button>
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
        <button class="btn btn-primary join-item" @click="getCode" :disabled="code?.deleted">
          Get a code
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
