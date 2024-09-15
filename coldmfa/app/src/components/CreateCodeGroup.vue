<script setup lang="ts">
import type { AxiosError, AxiosInstance, AxiosResponse } from 'axios'
import { ref } from 'vue'
import type { ApiError, CodeGroup } from '@/types'

const props = defineProps<{
  client: AxiosInstance
}>()

const emit = defineEmits<{
  created: [group: CodeGroup]
}>()

const groupName = ref('')
const errMsg = ref('')

const createGroup = (event: Event) => {
  event.preventDefault()
  event.stopPropagation()

  props.client
    .post('api/groups', {
      name: groupName.value
    })
    .then((response: AxiosResponse<CodeGroup | ApiError>) => {
      if (response.status === 201) {
        // TODO update state store
        groupName.value = ''
        errMsg.value = ''
        emit('created', response.data as CodeGroup)
      } else {
        // TODO handle auth error which doesn't return an error message
        errMsg.value = (response.data as ApiError).error
      }
    })
    .catch((err: AxiosError) => {
      console.error(err)
    })
}
</script>

<template>
  <p class="text-xl bold">Create a new group</p>
  <form class="flex flex-col" @submit="createGroup">
    <p v-if="errMsg">{{ errMsg }}</p>
    <label for="group">Group</label>
    <input type="text" id="group" class="rounded" v-model="groupName" />
    <div class="flex flex-row justify-end my-5 mx-1">
      <button class="bg-teal-400 rounded p-2 mt-2" type="submit">Group</button>
    </div>
  </form>
</template>
