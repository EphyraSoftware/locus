import { acceptHMRUpdate, defineStore } from 'pinia'
import { ref } from 'vue'
import type { CodeGroup, CodeSummary } from '@/types'

export const useGroupsStore = defineStore('groups', () => {
  const groups = ref<CodeGroup[]>([])

  /**
   * Insert a group into the store. If the group already exists, the existing group will be replaced.
   *
   * @param group The group to insert
   */
  const insertGroup = (group: CodeGroup) => {
    const existingIndex = groups.value.findIndex((g) => g.groupId === group.groupId)
    if (existingIndex !== -1) {
      groups.value[existingIndex] = group
    } else {
      groups.value.push(group)
    }
  }

  const setGroups = (newGroups: CodeGroup[]) => {
    groups.value = newGroups
  }

  const addCodeToGroup = (groupId: string, code: CodeSummary) => {
    const group = groups.value.find((g) => g.groupId === groupId)
    if (group) {
      if (!group.codes) {
        group.codes = []
      }
      group.codes.push(code)
    }
  }

  const groupHasCodes = (groupId: string) => {
    const group = groups.value.find((g) => g.groupId === groupId)
    return group && group.codes && group.codes.length > 0
  }

  const groupById = (groupId: string) => {
    return groups.value.find((g) => g.groupId === groupId)
  }

  return { groups, insertGroup, setGroups, addCodeToGroup, groupHasCodes, groupById }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useGroupsStore, import.meta.hot))
}
