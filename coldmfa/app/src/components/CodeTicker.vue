<script setup lang="ts">
import type { PasscodeResponse } from '@/types'
import { computed, inject, type Ref, ref, watch } from 'vue'

const props = defineProps<{
  passcodeResponse: PasscodeResponse
}>()

const emit = defineEmits<{
  expired: [serverTime: number]
}>()

const clientClock = inject('clientClock') as Ref<number>
const showCode = ref('')
const expired = ref(false)
const timerConfig = ref<TimerConfig>()

interface TimerConfig {
  windowStart: number
  windowEnd: number
  nextWindowEnd: number
}

watch(
  () => props.passcodeResponse,
  () => {
    // Compute bounds for the current and next windows
    const windowProgress = props.passcodeResponse.serverTime % props.passcodeResponse.period
    const windowStart = props.passcodeResponse.serverTime - windowProgress
    const windowEnd = windowStart + props.passcodeResponse.period
    const nextWindowEnd = windowEnd + props.passcodeResponse.period

    let newTimerConfig = {
      windowStart,
      windowEnd,
      nextWindowEnd
    } as TimerConfig

    showCode.value = props.passcodeResponse.passcode
    expired.value = false
    timerConfig.value = newTimerConfig
  },
  { immediate: true }
)

watch(clientClock, (currentTime) => {
  const timer = timerConfig.value
  if (expired.value || !timer) {
    return
  }

  if (currentTime <= timer.windowEnd) {
    // Nothing to do, set above
  } else if (currentTime <= timer.nextWindowEnd) {
    // Update to the next code, once
    if (showCode.value !== props.passcodeResponse.nextPasscode) {
      showCode.value = props.passcodeResponse.nextPasscode
    }
  } else {
    showCode.value = ''
    expired.value = true
    window.setTimeout(() => {
      emit('expired', props.passcodeResponse.serverTime)
    }, 3000)
  }
})

const clientSkew = computed(() => {
  return Math.abs(props.passcodeResponse.serverTime - Math.round(new Date().valueOf() / 1000))
})

const percentRemaining = () => {
  if (!timerConfig.value) {
    return 0
  }

  if (clientClock.value <= timerConfig.value.windowEnd) {
    return Math.floor(
      ((timerConfig.value.windowEnd - clientClock.value) / props.passcodeResponse.period) * 100
    )
  } else if (clientClock.value <= timerConfig.value.nextWindowEnd) {
    return Math.floor(
      ((timerConfig.value.nextWindowEnd - clientClock.value) / props.passcodeResponse.period) * 100
    )
  } else {
    return 0
  }
}

const secsRemaining = () => {
  if (!timerConfig.value) {
    return 0
  }

  if (clientClock.value <= timerConfig.value.windowEnd) {
    return timerConfig.value.windowEnd - clientClock.value
  } else if (clientClock.value <= timerConfig.value.nextWindowEnd) {
    return timerConfig.value.nextWindowEnd - clientClock.value
  } else {
    return 0
  }
}
</script>

<template>
  <p v-if="clientSkew > 1">Clock skew too significant</p>
  <p v-else-if="expired">Expired</p>
  <div v-else class="flex flex-row">
    <div class="flex flex-col justify-center px-3">
      <p>{{ showCode }}</p>
    </div>
    <div class="flex flex-col justify-center">
      <span
        class="radial-progress"
        :style="`--value:${percentRemaining()};--size:2rem;--thickness:2px;`"
        role="progressbar"
        >{{ secsRemaining() }}</span
      >
    </div>
  </div>
</template>

<style scoped></style>
