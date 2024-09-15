<script setup lang="ts">
import type { AxiosError, AxiosInstance, AxiosResponse } from 'axios'
import type { ApiError, CodeGroup } from '@/types'
import { ref } from 'vue'
import CreateCode from '@/components/CreateCode.vue'
import CodeList from '@/components/CodeList.vue'

const props = defineProps<{
  client: AxiosInstance
}>()

const groups = ref<CodeGroup[]>([])
const selectedGroupId = ref('')

props.client
  .get('api/groups')
  .then((response: AxiosResponse<CodeGroup[] | ApiError>) => {
    if (response.status === 200) {
      groups.value = response.data as CodeGroup[]
      if (groups.value.length > 0) {
        selectedGroupId.value = groups.value[0].group_id
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
  <select v-model="selectedGroupId">
    <option value="">Select a group</option>
    <option v-for="group in groups" :key="group.group_id" :value="group.group_id">
      {{ group.name }}
    </option>
  </select>

  <CreateCode :client="props.client" :groupId="selectedGroupId" />

  <CodeList :client="props.client" :group-id="selectedGroupId" />
</template>

<style scoped></style>
