<script setup lang="ts">
import type { AxiosInstance } from 'axios'
import type { CodeGroup } from '@/types'
import { computed, inject, watch } from 'vue'
import { useGroupsStore } from '@/stores/groups'

const props = defineProps<{
  groupId: string
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance
const groupsStore = useGroupsStore()

const codes = computed(() => {
  return groupsStore.groupById(props.groupId)?.codes ?? []
})

watch(
  () => props.groupId,
  (groupId) => {
    // Only load codes if we don't have details for this group yet.
    // Otherwise, maintain state on the UI.
    if (!groupsStore.groupHasCodes(groupId)) {
      fetchGroup(client, groupId)
    }
  }
)

const fetchGroup = async (client: AxiosInstance, groupId: string) => {
  if (groupId != '') {
    try {
      const response = await client.get(`api/groups/${groupId}`)

      if (response.status === 200 && (response.data as CodeGroup).codes) {
        let codeGroup = response.data as CodeGroup
        groupsStore.insertGroup(codeGroup)
      } else {
        console.error(response)
      }
    } catch (e) {
      console.error(e)
    }
  }
}
</script>

<template>
  <div class="flex flex-col">
    <p v-for="code in codes" :key="code.codeId">{{ code.preferredName ?? code.name }}</p>
  </div>
</template>

<style scoped></style>
