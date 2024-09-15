<script setup lang="ts">
import axios from 'axios'
import { provide, ref } from 'vue'
import CreateCodeGroup from '@/components/CreateCodeGroup.vue'
import CodeGroups from '@/components/CodeGroups.vue'

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
const showNewGroup = ref(false)

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

const groupCreated = () => {
  showNewGroup.value = false
}
</script>

<template>
  <header class="container mx-auto flex justify-center my-5">
    <div class="wrapper">
      <h1 class="bold text-3xl">Welcome to ColdMFA</h1>
    </div>
  </header>

  <main class="container mx-auto">
    <p>Welcome, {{ user }}</p>

    <div class="flex w-full justify-end">
      <button @click="showNewGroup = !showNewGroup" class="btn btn-secondary rounded p-2 mt-2">
        New group
      </button>
    </div>
    <div class="flex justify-center" v-if="showNewGroup">
      <div class="w-1/3">
        <CreateCodeGroup @created="groupCreated" />
      </div>
    </div>

    <CodeGroups />
  </main>
</template>

<style scoped></style>
