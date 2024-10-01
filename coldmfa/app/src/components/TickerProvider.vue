<script setup lang="ts">
import { onMounted, onUnmounted, provide, ref } from 'vue'

let timerInterval: number | undefined

const getTime = () => Math.round(new Date().valueOf() / 1000)

const clientClock = ref(getTime())
provide('clientClock', clientClock)

onMounted(() => {
  timerInterval = window.setInterval(() => {
    clientClock.value = getTime()
  }, 1000)
})

onUnmounted(() => {
  window.clearInterval(timerInterval)
})
</script>

<template>
  <slot></slot>
</template>

<style scoped></style>
