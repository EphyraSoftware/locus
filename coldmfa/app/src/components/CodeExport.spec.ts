import { describe, it, expect, beforeEach, beforeAll, afterAll, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { Pinia } from 'pinia'
import { createPinia, setActivePinia } from 'pinia'
import { useGroupsStore } from '@/stores/groups'
import axios, { type AxiosResponse } from 'axios'
import type { AxiosInstance } from 'axios'
import { resetHttpMockServer, startHttpMockServer, stopHttpMockServer } from '@/support/httpMock'
import type { CodeGroup, CodeSummary } from '@/types'
import CodeExport from '@/components/CodeExport.vue'
import { nextTick } from 'vue'

describe('CodeExport', () => {
  let client: AxiosInstance
  let pinia: Pinia
  let groupId: string

  beforeAll(() => {
    startHttpMockServer()
  })

  afterAll(() => {
    vi.useRealTimers()
    stopHttpMockServer()
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

    const groupResp: AxiosResponse<CodeGroup> = await client.post(
      'http://127.0.0.1:3000/coldmfa/api/groups',
      {
        name: 'Test Group'
      }
    )
    groupId = groupResp.data.groupId
    const groupsStore = useGroupsStore()
    groupsStore.insertGroup(groupResp.data)

    const codeResp: AxiosResponse<CodeSummary> = await client.post(
      `http://127.0.0.1:3000/coldmfa/api/groups/${groupId}/codes`,
      {
        original:
          'otpauth://totp/EphyraSoftware:test-a?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=NL6ZHWZXRNCNNIHQKDXK2Q4GGA3PKQD3'
      }
    )
    groupsStore.addCodeToGroup(groupId, codeResp.data)
  })

  it('render', async () => {
    const groupsStore = useGroupsStore()

    const wrapper = mount(CodeExport, {
      global: {
        provide: {
          client
        }
      },
      props: {
        groupId: groupsStore.groups[0].groupId,
        code: groupsStore.groups[0].codes![0]
      }
    })

    await flushPromises()

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')
  })

  it('auto hides', async () => {
    const groupsStore = useGroupsStore()

    vi.useFakeTimers({
      toFake: ['setTimeout']
    })

    const wrapper = mount(CodeExport, {
      global: {
        provide: {
          client
        }
      },
      props: {
        groupId: groupsStore.groups[0].groupId,
        code: groupsStore.groups[0].codes![0]
      }
    })

    await flushPromises()

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')

    vi.runAllTimers()
    await nextTick()

    const imgEl = wrapper.find('img')
    expect(imgEl.exists()).toBe(false)

    expect(wrapper.html()).toContain('Hidden...')
  })

  it('emit close', async () => {
    const groupsStore = useGroupsStore()

    const wrapper = mount(CodeExport, {
      global: {
        provide: {
          client
        }
      },
      props: {
        groupId: groupsStore.groups[0].groupId,
        code: groupsStore.groups[0].codes![0]
      }
    })

    await flushPromises()

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')

    await wrapper.get('button').trigger('click')

    expect('close' in wrapper.emitted()).toBe(true)
  })
})
