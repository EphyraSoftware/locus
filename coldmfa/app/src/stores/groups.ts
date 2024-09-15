import {defineStore} from "pinia";
import {ref} from "vue";
import type {CodeGroup} from "@/types";

export const useGroupsStore = defineStore('groups', () => {
    const groups = ref<CodeGroup[]>([])
    const insertGroup = (group: CodeGroup) => {
        groups.value.push(group)
    }
    const setGroups = (newGroups: CodeGroup[]) => {
        groups.value = newGroups
    }

    return {groups, insertGroup, setGroups}
})
