<script setup lang="ts">
import { inject, onMounted, ref, useTemplateRef } from 'vue'
import type { AxiosError, AxiosInstance, AxiosResponse } from 'axios'
import type { ApiError, CodeSummary } from '@/types'
import { useGroupsStore } from '@/stores/groups'

const props = defineProps<{
  groupId: string
}>()

const emit = defineEmits<{
  created: [code: CodeSummary]
}>()

const client = inject<AxiosInstance>('client') as AxiosInstance
const groupsStore = useGroupsStore()

const activeTab = ref<'url' | 'manual'>('url')
const input = useTemplateRef('codeNameInput')

// For URL input
const original = ref('')

// For manual input
const provider = ref('')
const codeName = ref('')
const secret = ref('')
const algorithm = ref('SHA1')
const digits = ref('6')
const period = ref('30')

// For error message
const errMsg = ref('')

onMounted(() => {
  input.value?.focus()
})

const storeCode = async () => {
  let codeUrl = original.value
  if (activeTab.value === 'manual') {
    codeUrl = `otpauth://totp/${provider.value}:${codeName.value}?algorithm=${algorithm.value}&digits=${digits.value}&issuer=${provider.value}&period=${period.value}&secret=${secret.value}`
  }

  try {
    const groupId = props.groupId
    const response: AxiosResponse<CodeSummary | ApiError> = await client.post(
      `api/groups/${groupId}/codes`,
      {
        original: codeUrl
      },
      {
        validateStatus: (status) => status === 201
      }
    )

    original.value = ''
    groupsStore.addCodeToGroup(groupId, response.data as CodeSummary)
    emit('created', response.data as CodeSummary)
  } catch (error) {
    const err = error as AxiosError
    if (
      err?.response &&
      err?.response?.data &&
      typeof err.response.data === 'object' &&
      'error' in err.response.data
    ) {
      errMsg.value = (err.response.data as ApiError).error
    } else {
      errMsg.value = 'Unknown error'
      console.error(err)
    }
  }
}
</script>

<template>
  <p class="text-xl bold">Create a new code</p>

  <div role="tablist" class="tabs tabs-boxed">
    <a
      role="tab"
      class="tab"
      :class="{ 'tab-active': activeTab === 'url' }"
      @click="activeTab = 'url'"
      >From URL</a
    >
    <a
      role="tab"
      class="tab"
      :class="{ 'tab-active': activeTab === 'manual' }"
      data-test-id="manual"
      @click="activeTab = 'manual'"
      >Manual</a
    >
  </div>

  <form class="flex flex-col py-2 gap-3" @submit.prevent.stop="storeCode">
    <template v-if="activeTab === 'url'">
      <input
        type="text"
        placeholder="URL for the One Time Password"
        name="original"
        class="input input-bordered input-accent w-full max-w-s"
        autocomplete="off"
        ref="codeNameInput"
        v-model="original"
        data-test-id="code-original"
      />
    </template>
    <template v-else-if="activeTab === 'manual'">
      <input
        type="text"
        placeholder="Provider"
        class="input input-bordered input-accent w-full max-w-s"
        v-model="provider"
        data-test-id="code-provider"
      />

      <input
        type="text"
        placeholder="Name"
        class="input input-bordered input-accent w-full max-w-s"
        v-model="codeName"
        data-test-id="code-name"
      />

      <input
        type="password"
        placeholder="Secret"
        class="input input-bordered input-accent w-full max-w-s"
        v-model="secret"
        data-test-id="code-secret"
      />

      <select
        v-model="algorithm"
        data-test-id="code-algorithm"
        class="select select-bordered w-full max-w-s"
      >
        <option value="SHA1">SHA1</option>
        <option value="SHA256">SHA256</option>
        <option value="SHA512">SHA512</option>
      </select>

      <select
        v-model="digits"
        data-test-id="code-digits"
        class="select select-bordered w-full max-w-s"
      >
        <option value="6">6</option>
        <option value="8">8</option>
      </select>

      <input
        type="number"
        placeholder="Period"
        class="input input-bordered input-accent w-full max-w-s"
        v-model="period"
        data-test-id="code-period"
      />
    </template>

    <p v-if="errMsg" class="text-red-500 py-2">Error creating your code: {{ errMsg }}</p>

    <div class="flex flex-row justify-end my-2 mx-1">
      <button class="btn btn-primary rounded p-2 mt-2" type="submit" data-test-id="create-code">
        Store code
      </button>
    </div>
  </form>
</template>

<style scoped></style>
