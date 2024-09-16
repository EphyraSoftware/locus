<script setup lang="ts">
import type { PasscodeResponse } from '@/types'
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  passcodeResponse: PasscodeResponse
}>()

const emit = defineEmits<{
  expired: [serverTime: number]
}>()

const clientTime = ref(0)
const showCode = ref('')
const expired = ref(false)
const timerConfig = ref<TimerConfig>()
const timerInterval = ref<number | undefined>()

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

    clientTime.value = Math.round(new Date().valueOf() / 1000)
    showCode.value = props.passcodeResponse.passcode
    expired.value = false

    // If there was a previous timer, clear it
    if (timerInterval.value) {
      clearInterval(timerInterval.value)
      timerInterval.value = undefined
    }

    const interval = setInterval(() => {
      let currentTimeMillis = new Date().valueOf()

      const currentTime = Math.round(currentTimeMillis / 1000)
      clientTime.value = currentTime
      if (currentTime <= newTimerConfig.windowEnd) {
        // Nothing to do, set above
      } else if (currentTime <= newTimerConfig.nextWindowEnd) {
        // Update to the next code, once
        if (showCode.value !== props.passcodeResponse.nextPasscode) {
          showCode.value = props.passcodeResponse.nextPasscode
        }
      } else {
        showCode.value = ''
        expired.value = true
        clearInterval(interval)
        setTimeout(() => {
          emit('expired', props.passcodeResponse.serverTime)
        }, 3000)
      }
    }, 1000)

    timerInterval.value = interval

    timerConfig.value = newTimerConfig
  },
  { immediate: true }
)

/*
const timerConfig = computed<TimerConfig>(() => {
  const windowProgress = props.passcodeResponse.serverTime % props.passcodeResponse.period
  const windowStart = props.passcodeResponse.serverTime - windowProgress
  const windowEnd = windowStart + props.passcodeResponse.period
  const nextWindowEnd = windowEnd + props.passcodeResponse.period

  let timerConfig = {
    windowStart,
    windowEnd,
    nextWindowEnd,
  } as TimerConfig;

  clientTime.value = Math.round(new Date().valueOf() / 1000)
  showCode.value = props.passcodeResponse.passcode
  expired.value = false

  if (timerInterval.value) {
    clearInterval(timerInterval.value)
    timerInterval.value = undefined
  }

  console.log('timerConfig', timerConfig)

  const interval = setInterval(() => {
    console.log('ticker')

    const currentTime = Math.round(new Date().valueOf() / 1000)
    clientTime.value = currentTime
    if (currentTime <= timerConfig.windowEnd) {
      // Nothing to do
    } else if (currentTime <= timerConfig.nextWindowEnd) {
      if (showCode.value !== props.passcodeResponse.nextPasscode) {
        showCode.value = props.passcodeResponse.nextPasscode
      }
    } else {
      expired.value = true
      clearInterval(interval)
    }
  }, 1000)

  setTimeout(() => {
    clearInterval(interval)
  }, (timerConfig.nextWindowEnd - clientTime.value) * 1000)

  timerInterval.value = interval

  return timerConfig
})
*/

const clientSkew = computed(() => {
  return Math.abs(props.passcodeResponse.serverTime - Math.round(new Date().valueOf() / 1000))
})

const percentRemaining = () => {
  if (clientTime.value <= timerConfig.value.windowEnd) {
    return Math.floor(
      ((timerConfig.value.windowEnd - clientTime.value) / props.passcodeResponse.period) * 100
    )
  } else if (clientTime.value <= timerConfig.value.nextWindowEnd) {
    return Math.floor(
      ((timerConfig.value.nextWindowEnd - clientTime.value) / props.passcodeResponse.period) * 100
    )
  } else {
    return 0
  }
}

const secsRemaining = () => {
  if (clientTime.value <= timerConfig.value.windowEnd) {
    return timerConfig.value.windowEnd - clientTime.value
  } else if (clientTime.value <= timerConfig.value.nextWindowEnd) {
    return timerConfig.value.nextWindowEnd - clientTime.value
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
