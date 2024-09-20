<script setup lang="ts">
import type { AxiosError, AxiosInstance, AxiosResponse } from 'axios'
import { inject, onMounted, ref, useTemplateRef } from 'vue'
import type { ApiError, CodeGroup } from '@/types'
import { useGroupsStore } from '@/stores/groups'

const emit = defineEmits<{
  created: [group: CodeGroup]
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance
const groupsStore = useGroupsStore()

const groupName = ref('')
const errMsg = ref('')

const input = useTemplateRef('groupNameInput')

onMounted(() => {
  input.value?.focus()
})

const createGroup = (event: Event) => {
  event.preventDefault()
  event.stopPropagation()

  client
    .post('api/groups', {
      name: groupName.value
    })
    .then((response: AxiosResponse<CodeGroup | ApiError>) => {
      if (response.status === 201) {
        groupsStore.insertGroup(response.data as CodeGroup)
        groupName.value = ''
        errMsg.value = ''
        emit('created', response.data as CodeGroup)
      } else {
        // TODO handle auth error which doesn't return an error message
        errMsg.value = (response.data as ApiError).error
      }
    })
    .catch((err: AxiosError) => {
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
    })
}
</script>

<template>
  <p class="text-xl bold">Create a new group</p>
  <form class="flex flex-col py-2" @submit="createGroup">
    <input
      type="text"
      placeholder="Group name"
      class="input input-bordered input-accent w-full max-w-s"
      autocomplete="off"
      ref="groupNameInput"
      v-model="groupName"
    />
    <p v-if="errMsg" class="text-red-500 py-2">Error creating your group: {{ errMsg }}</p>

    <div class="flex flex-row justify-end my-2 mx-1">
      <button class="btn btn-primary rounded p-2 mt-2" type="submit">Create</button>
    </div>
  </form>
</template>
