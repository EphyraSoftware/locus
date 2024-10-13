import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import type { AxiosInstance, AxiosResponse } from 'axios'
import type { Pinia } from 'pinia'
import { resetHttpMockServer, startHttpMockServer, stopHttpMockServer } from '@/support/httpMock'
import axios from 'axios'
import { createPinia, setActivePinia } from 'pinia'
import type { CodeGroup, CodeSummary } from '@/types'
import { useGroupsStore } from '@/stores/groups'
import CodeSummaryLine from '@/components/CodeSummaryLine.vue'
import TickerProvider from '@/components/TickerProvider.vue'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'

describe('CodeSummaryLine', () => {
  let client: AxiosInstance
  let pinia: Pinia
  let groupId: string
  let otherGroupId: string
  let codeId: string

  const Host = {
    name: 'Host',
    template: `
          <TickerProvider>
            <CodeSummaryLine :group-id="this.groupId" :code-id="this.codeId" :show-name-update-button="true" @show-export="(c) => $emit('showExport', c)" />
          </TickerProvider>
        `,
    props: ['groupId', 'codeId'],
    emits: ['showExport'],
    components: {
      TickerProvider,
      CodeSummaryLine
    }
  }

  beforeAll(() => {
    startHttpMockServer()
    vi.useFakeTimers()
  })

  afterAll(() => {
    stopHttpMockServer()
    vi.useRealTimers()
  })

  beforeEach(async () => {
    resetHttpMockServer()

    client = axios.create({
      baseURL: 'http://127.0.0.1:3000/coldmfa',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json'
      },
      withCredentials: true
    })

    pinia = createPinia()
    setActivePinia(pinia)

    const resp: AxiosResponse<CodeGroup> = await client.post(
      'http://127.0.0.1:3000/coldmfa/api/groups',
      {
        name: 'Test Group'
      }
    )
    groupId = resp.data.groupId
    const groupsStore = useGroupsStore()
    groupsStore.insertGroup(resp.data)

    const otherGroupResp: AxiosResponse<CodeGroup> = await client.post(
      'http://127.0.0.1:3000/coldmfa/api/groups',
      {
        name: 'Other Test Group'
      }
    )
    otherGroupId = otherGroupResp.data.groupId
    groupsStore.insertGroup(otherGroupResp.data)

    const codeResp: AxiosResponse<CodeSummary> = await client.post(
      `http://127.0.0.1:3000/coldmfa/api/groups/${groupId}/codes`,
      {
        original:
          'otpauth://totp/EphyraSoftware:test-a?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=NL6ZHWZXRNCNNIHQKDXK2Q4GGA3PKQD3'
      }
    )
    codeId = codeResp.data.codeId
    groupsStore.addCodeToGroup(groupId, codeResp.data)
  })

  it('render', () => {
    const wrapper = mount(Host, {
      props: {
        groupId,
        codeId
      },
      global: {
        provide: {
          client
        }
      }
    })

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')
  })

  it('get a code', async () => {
    const wrapper = mount(Host, {
      props: {
        groupId,
        codeId
      },
      global: {
        provide: {
          client
        }
      }
    })

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')

    await wrapper.get('button[data-test-id="get-code"]').trigger('click')
    await flushPromises()

    expect(wrapper.html()).toContain('123456')
  })

  it('remove expired codes', async () => {
    const wrapper = mount(Host, {
      props: {
        groupId,
        codeId
      },
      global: {
        provide: {
          client
        }
      }
    })

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')

    await wrapper.get('button[data-test-id="get-code"]').trigger('click')
    await flushPromises()

    expect(wrapper.html()).toContain('123456')

    for (let i = 0; i < 61; i++) {
      vi.advanceTimersToNextTimer()
      await nextTick()

      if (wrapper.html().includes('Expired')) {
        break
      }
    }

    expect(wrapper.html()).toContain('Expired')

    for (let i = 0; i < 3; i++) {
      vi.advanceTimersToNextTimer()
      await nextTick()
    }

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')
    expect(wrapper.html()).not.toContain('Expired')
  })

  it('rename a code', async () => {
    const wrapper = mount(Host, {
      props: {
        groupId,
        codeId
      },
      global: {
        provide: {
          client
        }
      }
    })

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')

    wrapper.get('p[data-test-id="code-name"]').element.textContent = 'my-test-code'

    // Click the rename button (only visible in testing)
    await wrapper.get('button[data-test-id="rename"]').trigger('click')

    await flushPromises()

    expect(wrapper.html()).toContain('my-test-code')

    const groupsStore = useGroupsStore()
    const code = groupsStore.codeById(groupId, codeId)
    expect(code?.preferredName).toBe('my-test-code')
  })

  it('request export', async () => {
    const wrapper = mount(Host, {
      props: {
        groupId,
        codeId
      },
      global: {
        provide: {
          client
        }
      }
    })

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')

    await wrapper.get('button[data-test-id="export"]').trigger('click')

    expect('showExport' in wrapper.emitted()).toBe(true)

    const exportCodeId = wrapper.emitted()['showExport'] as string[][]
    expect(exportCodeId[0][0]).toEqual(codeId)
  })

  it('delete', async () => {
    const wrapper = mount(Host, {
      props: {
        groupId,
        codeId
      },
      global: {
        provide: {
          client
        }
      }
    })

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')

    for (let i = 0; i < 5; i++) {
      await wrapper.get('button[data-test-id="delete"]').trigger('click')
    }

    await flushPromises()

    const groupsStore = useGroupsStore()
    const code = groupsStore.codeById(groupId, codeId)
    expect(code?.deleted).toBe(true)
  })

  it('move between groups', async () => {
    const wrapper = mount(Host, {
      props: {
        groupId,
        codeId
      },
      global: {
        provide: {
          client
        }
      }
    })

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')

    await wrapper.get('button[data-test-id="move"]').trigger('click')

    const otherGroups = wrapper.findAll('li')
    expect(otherGroups.length).toBe(1)
    expect(otherGroups[0].html()).toContain('Other Test Group')
    expect(otherGroups[0].isVisible()).toBe(true)

    // Select the group to switch to and click it
    // That triggers a move request to the backend and should update the store
    await otherGroups[0].trigger('click')
    await flushPromises()

    expect(wrapper.html()).not.toContain('EphyraSoftware:test-a')

    // Switch to the other group, where the code should have moved to
    await wrapper.setProps({
      groupId: otherGroupId
    })

    // Now it's showing again
    expect(wrapper.html()).toContain('EphyraSoftware:test-a')

    // Check the state was correctly updated in the store
    const groupsStore = useGroupsStore()
    const removedCode = groupsStore.codeById(groupId, codeId)
    expect(removedCode).toBe(undefined)
    const code = groupsStore.codeById(otherGroupId, codeId)
    expect(code?.codeId).toBe(codeId)
  })
})
