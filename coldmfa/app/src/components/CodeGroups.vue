<script setup lang="ts">
import type { AxiosError, AxiosInstance, AxiosResponse } from 'axios'
import type { ApiError, CodeGroup } from '@/types'
import { inject, ref } from 'vue'
import CreateCode from '@/components/CreateCode.vue'
import CodeList from '@/components/CodeList.vue'
import { useGroupsStore } from '@/stores/groups'

const client = inject<AxiosInstance>('client') as AxiosInstance
const groupsStore = useGroupsStore()
const selectedGroupId = ref('')
const showNewCode = ref(false)

groupsStore.$subscribe((_, state) => {
  if (selectedGroupId.value === '' && state.groups.length > 0) {
    selectedGroupId.value = state.groups[0].groupId
  }
})

client
  .get('api/groups')
  .then((response: AxiosResponse<CodeGroup[] | ApiError>) => {
    if (response.status === 200) {
      groupsStore.setGroups(response.data as CodeGroup[])
      if (groupsStore.groups.length > 0) {
        selectedGroupId.value = groupsStore.groups[0].groupId
      }
    } else {
      console.error(response.data)
    }
  })
  .catch((err: AxiosError) => {
    console.error(err)
  })
</script>

<template>
  <div class="mb-5">
    <select
      v-model="selectedGroupId"
      :disabled="groupsStore.groups.length === 0"
      class="select select-bordered w-full max-w-xs"
    >
      <option disabled value="">Select a group</option>
      <option v-for="group in groupsStore.groups" :key="group.groupId" :value="group.groupId">
        {{ group.name }}
      </option>
    </select>
  </div>

  <div class="flex w-full justify-end">
    <button
      @click="showNewCode = !showNewCode"
      :disabled="selectedGroupId === ''"
      class="btn btn-secondary rounded p-2 mt-2"
    >
      New code
    </button>
  </div>
  <div class="flex justify-center" v-if="showNewCode">
    <div class="flex flex-col w-1/3">
      <CreateCode :group-id="selectedGroupId" @created="showNewCode = false" />
    </div>
  </div>

  <div class="my-3">
    <CodeList :group-id="selectedGroupId" :show-update-name-button="false" />
  </div>
</template>

<style scoped></style>
