<script setup lang="ts">
import axios from 'axios'
import { provide, ref } from 'vue'
import CreateCodeGroup from '@/components/CreateCodeGroup.vue'
import CodeGroups from '@/components/CodeGroups.vue'
import CreateBackup from '@/components/CreateBackup.vue'
import RestoreBackup from '@/components/RestoreBackup.vue'
import type { BackupWarning } from '@/types'

interface UserName {
  username: string
}

interface UserDetails {
  email: string
  name: UserName
}

interface User {
  user: UserDetails
}

const user = ref('')
const lastBackup = ref<string>()
const numberNotBackedUp = ref<number>()
const showNewGroup = ref(false)
const showBackup = ref(false)
const showRestore = ref(false)

const client = axios.create({
  baseURL: import.meta.env.DEV ? 'http://127.0.0.1:3000/coldmfa' : import.meta.env.BASE_URL,
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json'
  },
  withCredentials: true
})

provide('client', client)

client.get('api/user').then((response) => {
  const u = response.data as User
  user.value = u.user.name.username
})

client
  .get('api/backups/warning', { validateStatus: (s) => s === 200 })
  .then((response) => {
    const warning = response.data as BackupWarning
    if (warning.lastBackupAt) {
      lastBackup.value = new Date(warning.lastBackupAt).toString()
    } else {
      lastBackup.value = 'Never'
    }
    console.log(warning)
    numberNotBackedUp.value = warning.numberNotBackedUp
  })
  .catch((e) => {
    console.error(e)
  })

const showCreateGroup = () => {
  showNewGroup.value = !showNewGroup.value
  showBackup.value = false
  showRestore.value = false
}

const showCreateBackup = () => {
  showBackup.value = !showBackup.value
  showNewGroup.value = false
  showRestore.value = false
}

const showRestoreBackup = () => {
  showRestore.value = !showRestore.value
  showBackup.value = false
  showNewGroup.value = false
}

const groupCreated = () => {
  showNewGroup.value = false
}

const backupCompleted = () => {
  showBackup.value = false
}
</script>

<template>
  <header class="container mx-auto flex justify-center my-5">
    <div class="wrapper">
      <h1 class="bold text-3xl">Welcome to ColdMFA</h1>
    </div>
  </header>

  <main class="container mx-auto">
    <div class="flex flex-row">
      <div class="w-1/2">
        <p>Welcome, {{ user }}</p>
      </div>
      <div class="flex justify-end w-1/2">
        <p class="text-right">
          <span v-if="lastBackup">Last backup at: {{ lastBackup }}</span
          ><br />
          <span v-if="numberNotBackedUp && numberNotBackedUp > 0"
            >You have {{ numberNotBackedUp }} code needing backup</span
          >
        </p>
      </div>
    </div>

    <div class="flex w-full justify-end">
      <div class="join p-2 mt-2">
        <button @click="showCreateGroup" class="btn btn-secondary join-item">New group</button>
        <button @click="showCreateBackup" class="btn btn-secondary join-item">Backup</button>
        <button @click="showRestoreBackup" class="btn btn-secondary join-item">Restore</button>
      </div>
    </div>
    <div class="flex justify-center">
      <div class="w-1/3" v-if="showNewGroup">
        <CreateCodeGroup @created="groupCreated" />
      </div>
      <div class="w-1/3" v-if="showBackup">
        <CreateBackup @completed="backupCompleted" />
      </div>
      <div class="w-1/3" v-if="showRestore">
        <RestoreBackup />
      </div>
    </div>

    <CodeGroups />
  </main>
</template>

<style scoped></style>
