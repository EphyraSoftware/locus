import { describe, it, vi, expect, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import TickerProvider from '@/components/TickerProvider.vue'
import { type Component, nextTick } from 'vue'

describe('TickerProvider', () => {
  const Inner = {
    name: 'Inner',
    template: `
            <p>Current time: {{ clientClock }}</p>
        `,
    inject: ['clientClock']
  } as Component

  const Host = {
    name: 'Host',
    template: `
            <TickerProvider>
                <Inner />
            </TickerProvider>
        `,
    components: {
      TickerProvider,
      Inner
    }
  } as Component

  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('render', async () => {
    const wrapper = mount(Host)
    expect(wrapper.html()).toContain('Current time')
  })

  it('tick', async () => {
    vi.setSystemTime(0)
    const wrapper = mount(Host)

    expect(wrapper.html()).toContain('Current time: 0')

    vi.advanceTimersToNextTimer()
    await nextTick()

    expect(wrapper.html()).toContain('Current time: 1')

    vi.advanceTimersToNextTimer()
    await nextTick()

    expect(wrapper.html()).toContain('Current time: 2')
  })
})
