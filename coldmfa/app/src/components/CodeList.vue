<script setup lang="ts">
import type { AxiosInstance } from 'axios'
import type { CodeGroup, CodeSummary } from '@/types'
import { ref, watch } from 'vue'

const props = defineProps<{
  client: AxiosInstance
  groupId: string
}>()

const codes = ref<CodeSummary[]>([])

watch(
  () => props.groupId,
  (groupId) => {
    fetchGroup(props.client, groupId).then((codeSummaryList) => {
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
    <p v-for="code in codes" :key="code.code_id">{{ code.preferred_name ?? code.name }}</p>
  </div>
</template>

<style scoped></style>
