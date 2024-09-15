<script setup lang="ts">
import type { AxiosInstance } from 'axios'
import type { CodeGroup, CodeSummary } from '@/types'
import {inject, ref, watch} from 'vue'

const props = defineProps<{
  groupId: string
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance

const codes = ref<CodeSummary[]>([])

watch(
  () => props.groupId,
  (groupId) => {
    fetchGroup(client, groupId).then((codeSummaryList) => {
      if (codeSummaryList) {
        codes.value = codeSummaryList
      }
    })
  }
)

const fetchGroup = async (client: AxiosInstance, groupId: string) => {
  if (groupId != '') {
    try {
      const response = await client.get(`api/groups/${groupId}`)

      if (response.status === 200 && (response.data as CodeGroup).codes) {
        return (response.data as CodeGroup).codes
      } else {
        console.error(response)
      }
    } catch (e) {
      console.error(e)
    }
  }

  return []
}
</script>

<template>
  <div class="flex flex-col">
    <p v-for="code in codes" :key="code.codeId">{{ code.preferredName ?? code.name }}</p>
  </div>
</template>

<style scoped></style>
