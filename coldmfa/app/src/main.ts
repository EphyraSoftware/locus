import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import axios from 'axios'

const client = axios.create({
  baseURL: import.meta.env.DEV ? 'http://127.0.0.1:3000/coldmfa' : import.meta.env.BASE_URL,
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json'
  },
  withCredentials: true
})

const app = createApp(App, {
  client
})

app.use(createPinia())

app.mount('#app')
