import { afterAll, beforeAll, beforeEach, describe, it, expect } from 'vitest'
import { flushPromises, mount, VueWrapper } from '@vue/test-utils'
import type { AxiosInstance, AxiosResponse } from 'axios'
import type { Pinia } from 'pinia'
import { resetHttpMockServer, startHttpMockServer, stopHttpMockServer } from '@/support/httpMock'
import axios from 'axios'
import { createPinia, setActivePinia } from 'pinia'
import type { CodeGroup } from '@/types'
import { useGroupsStore } from '@/stores/groups'
import CodeGroups from '@/components/CodeGroups.vue'

describe('CodeGroups', () => {
  let client: AxiosInstance
  let pinia: Pinia
  let groupId1: string
  let groupId2: string

  beforeAll(() => {
    startHttpMockServer()
  })

  afterAll(() => {
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

    const groupsStore = useGroupsStore()

    const resp1: AxiosResponse<CodeGroup> = await client.post(
      'http://127.0.0.1:3000/coldmfa/api/groups',
      {
        name: 'Test Group 1'
      }
    )
    groupId1 = resp1.data.groupId
    groupsStore.insertGroup(resp1.data)

    const resp2: AxiosResponse<CodeGroup> = await client.post(
      'http://127.0.0.1:3000/coldmfa/api/groups',
      {
        name: 'Test Group 2'
      }
    )
    groupId2 = resp2.data.groupId
    groupsStore.insertGroup(resp2.data)
  })

  const createCode = async (wrapper: VueWrapper, original: string) => {
    await wrapper.get('button[data-test-id="new-code"]').trigger('click')

    await wrapper.get('input[data-test-id="code-original"]').setValue(original)
    await wrapper.get('button[data-test-id="create-code"]').trigger('submit')
    await flushPromises()
  }

  it('render', async () => {
    mount(CodeGroups, {
      global: {
        provide: {
          client
        }
      }
    })

    await flushPromises()
  })

  it('show/hide create a code', async () => {
    const wrapper = mount(CodeGroups, {
      global: {
        provide: {
          client
        }
      }
    })

    await flushPromises()

    await wrapper.get('button[data-test-id="new-code"]').trigger('click')

    expect(wrapper.find('button[data-test-id="create-code"]').exists()).toBe(true)

    await wrapper.get('button[data-test-id="new-code"]').trigger('click')

    expect(wrapper.find('button[data-test-id="create-code"]').exists()).toBe(false)
  })

  it('create a code', async () => {
    const wrapper = mount(CodeGroups, {
      global: {
        provide: {
          client
        }
      }
    })

    await flushPromises()

    await wrapper.get('button[data-test-id="new-code"]').trigger('click')

    await wrapper
      .get('input[data-test-id="code-original"]')
      .setValue(
        'otpauth://totp/EphyraSoftware:test-a?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=NL6ZHWZXRNCNNIHQKDXK2Q4GGA3PKQD3'
      )
    await wrapper.get('button[data-test-id="create-code"]').trigger('submit')
    await flushPromises()

    expect(wrapper.find('button[data-test-id="create-code"]').exists()).toBe(false)

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')
  })

  it('create a code in each', async () => {
    const wrapper = mount(CodeGroups, {
      global: {
        provide: {
          client
        }
      }
    })

    await flushPromises()

    await createCode(
      wrapper,
      'otpauth://totp/EphyraSoftware:test-a?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=NL6ZHWZXRNCNNIHQKDXK2Q4GGA3PKQD3'
    )

    expect(wrapper.find('button[data-test-id="create-code"]').exists()).toBe(false)

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')

    await wrapper.get('select[data-test-id="group-select"]').setValue(groupId2)

    expect(wrapper.html()).not.toContain('EphyraSoftware:test-a')

    await createCode(
      wrapper,
      'otpauth://totp/EphyraSoftware:test-b?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=A23DJ4WDRR2XFPDKBUQ5ZLZN6KVIIIC4'
    )

    expect(wrapper.html()).toContain('EphyraSoftware:test-b')

    await wrapper.get('select[data-test-id="group-select"]').setValue(groupId1)

    expect(wrapper.html()).not.toContain('EphyraSoftware:test-b')
    expect(wrapper.html()).toContain('EphyraSoftware:test-a')
  })
})
