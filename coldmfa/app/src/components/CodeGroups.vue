<script setup lang="ts">
import type {AxiosError, AxiosInstance, AxiosResponse} from "axios";
import type {ApiError, CodeGroup} from "@/types";
import {ref} from "vue";

const props = defineProps<{
  client: AxiosInstance
}>()

const groups = ref<CodeGroup[]>([])
const selectedGroupId = ref('')

props.client.get("api/groups").then((response: AxiosResponse<CodeGroup[] | ApiError>) => {
  if (response.status === 200) {
    groups.value = response.data as CodeGroup[]
    if (groups.value.length > 0) {
      selectedGroupId.value = groups.value[0].id
    }
  } else {
    console.error(response.data)
  }
}).catch((err: AxiosError) => {
  console.error(err)
})
</script>

<template>
  <select v-model="selectedGroupId">
    <option value="">Select a group</option>
    <option v-for="group in groups" :key="group.id" :value="group.id">{{ group.name }}</option>
  </select>
</template>

<style scoped>

</style>