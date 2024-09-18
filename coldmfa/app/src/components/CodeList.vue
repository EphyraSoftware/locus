<script setup lang="ts">
import type { AxiosInstance } from 'axios'
import type { CodeGroup, CodeSummary } from '@/types'
import { computed, inject, ref, watch } from 'vue'
import { useGroupsStore } from '@/stores/groups'
import CodeSummaryLine from '@/components/CodeSummaryLine.vue'
import CodeExport from '@/components/CodeExport.vue'

const props = defineProps<{
  groupId: string
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance
const groupsStore = useGroupsStore()
const showExportFor = ref<CodeSummary>()

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
  <div class="flex justify-center">
    <template v-if="showExportFor">
      <CodeExport
        :group-id="props.groupId"
        :code="showExportFor"
        @close="showExportFor = undefined"
      ></CodeExport>
    </template>
  </div>

  <div class="flex flex-col">
    <div v-for="code in codes" :key="code.codeId" class="my-2">
      <CodeSummaryLine
        :group-id="props.groupId"
        :code-id="code.codeId"
        @show-export="
          (codeId) => {
            showExportFor = codes.find((c) => c.codeId === codeId)
          }
        "
        >{{ code.preferredName ?? code.name }}
      </CodeSummaryLine>
    </div>
  </div>
</template>

<style scoped></style>
