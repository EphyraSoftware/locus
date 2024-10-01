import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import TickerProvider from "@/components/TickerProvider.vue";
import type {Component} from "vue";
import CodeTicker from "@/components/CodeTicker.vue";
import type {PasscodeResponse} from "@/types";
import {mount} from "@vue/test-utils";
import {nextTick} from "vue";

describe('CodeTicker', () => {
    const Host = {
        name: 'Host',
        template: `
            <TickerProvider>
                <CodeTicker :passcode-response="this.response" @expired="$emit('expired')" />
            </TickerProvider>
        `,
        props: ['response'],
        emits: ['expired'],
        components: {
            TickerProvider,
            CodeTicker
        }
    } as Component

    beforeEach(() => {
        vi.useFakeTimers()
    })

    afterEach(() => {
        vi.useRealTimers()
    })

    it('render', async () => {
        vi.setSystemTime(0)
        const wrapper = mount(Host, {
            props: {
                response: {
                    serverTime: 0,
                    passcode: '123456',
                    nextPasscode: '654321',
                    period: 30
                } as PasscodeResponse
            }
        })
        expect(wrapper.html()).toContain('123456')
    })

    it('transition code', async () => {
        vi.setSystemTime(89_000)
        const wrapper = mount(Host, {
            props: {
                response: {
                    serverTime: 89,
                    passcode: '123456',
                    nextPasscode: '654321',
                    period: 30
                } as PasscodeResponse
            }
        })

        expect(wrapper.html()).toContain('123456')

        for (let i = 0; i < 2; i++) {
            vi.advanceTimersToNextTimer()
            await nextTick()
        }

        expect(wrapper.html()).toContain('654321')
    })

    it('expire codes', async () => {
        vi.setSystemTime(89_000)
        const wrapper = mount(Host, {
            props: {
                response: {
                    serverTime: 89,
                    passcode: '123456',
                    nextPasscode: '654321',
                    period: 30
                } as PasscodeResponse
            }
        })

        expect(wrapper.html()).toContain('123456')

        for (let i = 0; i < 2; i++) {
            vi.advanceTimersToNextTimer()
            await nextTick()
        }

        expect(wrapper.html()).toContain('654321')

        for (let i = 0; i < 30; i++) {
            vi.advanceTimersToNextTimer()
            await nextTick()
        }

        expect(wrapper.html()).toContain('Expired')
        expect('expired' in wrapper.emitted()).toBe(false)

        vi.advanceTimersToNextTimer()
        await nextTick()

        expect(wrapper.html()).toContain('Expired')

        // The clock ticker is still running, so we have to advance 3 times to move 3 seconds
        // which is how long the Expired message shows for
        for (let i = 0; i < 3; i++) {
            vi.advanceTimersToNextTimer()
            await nextTick()
        }

        expect('expired' in wrapper.emitted()).toBe(true)
    })

    it('rejects clock skew', async () => {
        vi.setSystemTime(100_000)
        const wrapper = mount(Host, {
            props: {
                response: {
                    serverTime: 105,
                    passcode: '123456',
                    nextPasscode: '654321',
                    period: 30
                } as PasscodeResponse
            }
        })

        expect(wrapper.html()).toContain('Clock skew too significant')
    })
})
