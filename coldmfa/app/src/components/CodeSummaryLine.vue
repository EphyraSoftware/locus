<script setup lang="ts">
import type { ApiError, PasscodeResponse } from '@/types'
import type { AxiosInstance, AxiosResponse } from 'axios'
import { computed, inject, ref, useTemplateRef } from 'vue'
import CodeTicker from '@/components/CodeTicker.vue'
import { useGroupsStore } from '@/stores/groups'

const props = defineProps<{
  groupId: string
  codeId: string
  showNameUpdateButton: boolean
}>()

defineEmits<{
  showExport: [codeId: string]
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance
const groupsStore = useGroupsStore()
const fetchedCode = ref<PasscodeResponse | undefined>()
const moveCodeModalOpen = ref(false)

const codeName = useTemplateRef<HTMLParagraphElement>('codeName')

const code = computed(() => groupsStore.codeById(props.groupId, props.codeId))
const otherGroups = computed(() => groupsStore.groups.filter((g) => g.groupId !== props.groupId))

const deleteCounter = ref(5)

const getCode = async () => {
  try {
    const codesResponse: AxiosResponse<PasscodeResponse> = await client.get(
      `api/groups/${props.groupId}/codes/${props.codeId}`,
      {
        validateStatus: (status) => status === 200
      }
    )

    fetchedCode.value = codesResponse.data
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
    if (newPreferredName === null || newPreferredName === currentCode?.preferredName) {
      return
    }

    await client.put(
      `api/groups/${props.groupId}/codes/${currentCode.codeId}`,
      {
        preferredName: newPreferredName
      },
      {
        validateStatus: (status) => status === 204
      }
    )

    if (newPreferredName === '') {
      currentCode.preferredName = undefined
    } else {
      currentCode.preferredName = newPreferredName
    }
    groupsStore.replaceCodeInGroup(props.groupId, currentCode)
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

  deleteCounter.value--

  if (deleteCounter.value <= 0) {
    try {
      await client.delete(`api/groups/${props.groupId}/codes/${props.codeId}`, {
        validateStatus: (status) => status === 204
      })

      currentCode.deleted = true
      currentCode.deletedAt = new Date().valueOf()
      groupsStore.replaceCodeInGroup(props.groupId, currentCode)
    } catch (e) {
      console.error(e)
    }
  }
}

const showMoveCode = () => {
  moveCodeModalOpen.value = true
}

const moveCode = async (groupId: string) => {
  if (!code.value || !groupId) {
    return
  }

  try {
    await client.post(
      `api/groups/${props.groupId}/codes/${code.value.codeId}/move`,
      {
        toGroupId: groupId
      },
      {
        validateStatus: (status) => status === 204
      }
    )

    groupsStore.addCodeToGroup(groupId, code.value)
    groupsStore.removeCodeFromGroup(props.groupId, code.value.codeId)
  } catch (e) {
    console.error(e)
  }

  moveCodeModalOpen.value = false
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
        data-test-id="code-name"
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
        <button
          class="btn btn-error join-item"
          @click="tryDelete"
          data-test-id="delete"
          :disabled="code?.deleted"
        >
          {{ deleteCounter == 5 ? 'Delete' : `Confirm? (${deleteCounter})` }}
        </button>
        <button
          class="btn btn-secondary join-item"
          data-test-id="export"
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
        <button class="btn btn-secondary join-item" data-test-id="move" @click="showMoveCode">
          Move
        </button>
        <button
          class="btn btn-primary join-item"
          @click="getCode"
          data-test-id="get-code"
          :disabled="code?.deleted"
        >
          Get a code
        </button>
        <button
          class="btn btn-primary join-item"
          @click="renameCode"
          data-test-id="rename"
          v-if="props.showNameUpdateButton"
        >
          Rename
        </button>
      </div>
    </div>
  </div>

  <dialog class="modal" :open="moveCodeModalOpen">
    <div class="modal-box">
      <h3 class="text-lg font-bold">Pick another group</h3>
      <div class="ms-5 mt-3">
        <ul class="list-disc">
          <li
            v-for="group in otherGroups"
            :key="group.groupId"
            class="cursor-pointer"
            @click="moveCode(group.groupId)"
          >
            {{ group.name }}
          </li>
        </ul>
      </div>

      <div class="modal-action">
        <form method="dialog">
          <!-- if there is a button in a form, it will close the modal -->
          <button class="btn">Cancel</button>
        </form>
      </div>
    </div>
  </dialog>
</template>

<style scoped></style>
