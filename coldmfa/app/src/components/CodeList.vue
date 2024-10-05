<script setup lang="ts">
import type { AxiosInstance } from 'axios'
import type { CodeGroup, CodeSummary } from '@/types'
import { computed, inject, ref, watch } from 'vue'
import { useGroupsStore } from '@/stores/groups'
import CodeSummaryLine from '@/components/CodeSummaryLine.vue'
import CodeExport from '@/components/CodeExport.vue'
import TickerProvider from '@/components/TickerProvider.vue'

const props = defineProps<{
  groupId: string
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance
const groupsStore = useGroupsStore()
const showExportFor = ref<CodeSummary>()

const sortBy = ref<'alpha' | 'create'>('alpha')
const showDeleted = ref(false)
const codes = computed(() => {
  let groupCodes = groupsStore.groupCodes(props.groupId)
  if (groupCodes) {
    groupCodes.sort((a, b) => {
      if (sortBy.value === 'alpha') {
        const aName = a.preferredName ?? a.name
        const bName = b.preferredName ?? b.name

        return aName.localeCompare(bName)
      } else {
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      }
    })

    if (!showDeleted.value) {
      groupCodes = groupCodes?.filter((code) => !code.deleted)
    }
  }

  return groupCodes
})

const fetchGroup = async (client: AxiosInstance, groupId: string) => {
  if (groupId != '') {
    const response = await client.get(`api/groups/${groupId}`)

    if (response.status === 200 && (response.data as CodeGroup).codes) {
      let codeGroup = response.data as CodeGroup
      groupsStore.insertGroup(codeGroup)
    } else {
      throw response
    }
  }
}

watch(
  () => props.groupId,
  (groupId) => {
    // Only load codes if we don't have details for this group yet.
    // Otherwise, maintain state on the UI.
    if (!groupsStore.groupHasCodes(groupId)) {
      try {
        fetchGroup(client, groupId)
      } catch (e) {
        console.error(e)
      }
    }
  },
  { immediate: true }
)
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

  <div class="flex flex-row mb-5">
    <select
      :disabled="!codes || codes.length === 0"
      class="select select-bordered w-full max-w-xs"
      v-model="sortBy"
    >
      <option value="alpha">Alphabetical</option>
      <option value="create">Creation date</option>
    </select>
    <div class="form-control ms-5">
      <label class="label cursor-pointer">
        <span class="pe-2 text-lg">Show deleted</span>
        <input type="checkbox" v-model="showDeleted" class="checkbox checkbox-md" />
      </label>
    </div>
  </div>

  <div class="flex flex-col">
    <div v-if="!codes || codes.length === 0" class="flex flex-row justify-center">
      <p>No codes yet</p>
    </div>
    <div v-else>
      <TickerProvider>
        <div v-for="code in codes" :key="code.codeId" class="my-2">
          <CodeSummaryLine
            :group-id="props.groupId"
            :code-id="code.codeId"
            :show-name-update-button="false"
            @show-export="
              (codeId) => {
                showExportFor = codes?.find((c) => c.codeId === codeId)
              }
            "
            >{{ code.preferredName ?? code.name }}
          </CodeSummaryLine>
        </div>
      </TickerProvider>
    </div>
  </div>
</template>

<style scoped></style>
