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

  const replaceCodeInGroup = (groupId: string, code: CodeSummary) => {
    const group = groups.value.find((g) => g.groupId === groupId)
    if (group && group.codes) {
      const codeIndex = group.codes.findIndex((c) => c.codeId === code.codeId)
      if (codeIndex !== -1) {
        group.codes[codeIndex] = code
      }
    }
  }

  const groupHasCodes = (groupId: string) => {
    const group = groups.value.find((g) => g.groupId === groupId)
    return group && group.codes && group.codes.length > 0
  }

  const groupById = (groupId: string) => {
    return groups.value.find((g) => g.groupId === groupId)
  }

  const codeById = (groupId: string, codeId: string) => {
    const group = groups.value.find((g) => g.groupId === groupId)
    return group && group.codes && group.codes.find((c) => c.codeId === codeId)
  }

  const groupCodes = (groupId: string) => {
    const group = groups.value.find((g) => g.groupId === groupId)
    return group && group.codes
  }

  const removeCodeFromGroup = (groupId: string, codeId: string) => {
    const group = groups.value.find((g) => g.groupId === groupId)
    if (group && group.codes) {
      const codeIndex = group.codes.findIndex((c) => c.codeId === codeId)
      if (codeIndex !== -1) {
        group.codes.splice(codeIndex, 1)
      }
    }
  }

  return {
    groups,
    insertGroup,
    setGroups,
    addCodeToGroup,
    replaceCodeInGroup,
    groupHasCodes,
    groupById,
    codeById,
    groupCodes,
    removeCodeFromGroup
  }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useGroupsStore, import.meta.hot))
}
