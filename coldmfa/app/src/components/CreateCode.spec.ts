import { describe, it, expect, beforeEach, beforeAll, afterAll } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { Pinia } from 'pinia'
import { createPinia, setActivePinia } from 'pinia'
import { useGroupsStore } from '@/stores/groups'
import axios, { type AxiosResponse } from 'axios'
import type { AxiosInstance } from 'axios'
import {
  resetHttpMockServer,
  setNextIsHttpErr,
  startHttpMockServer,
  stopHttpMockServer
} from '@/support/httpMock'
import CreateCode from '@/components/CreateCode.vue'
import type { CodeGroup } from '@/types'

describe('CreateCode', () => {
  let client: AxiosInstance
  let pinia: Pinia
  let groupId: string

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

    const resp: AxiosResponse<CodeGroup> = await client.post(
      'http://127.0.0.1:3000/coldmfa/api/groups',
      {
        name: 'Test Group'
      }
    )
    groupId = resp.data.groupId
    const groupsStore = useGroupsStore()
    groupsStore.insertGroup(resp.data)
  })

  it('render', () => {
    const wrapper = mount(CreateCode, {
      global: {
        provide: {
          client
        }
      },
      props: {
        groupId
      }
    })
    expect(wrapper.html()).toContain('Create a new code')
  })

  it('create a new code', async () => {
    const wrapper = mount(CreateCode, {
      global: {
        provide: {
          client
        }
      },
      props: {
        groupId
      }
    })
    await wrapper
      .get('input')
      .setValue(
        'otpauth://totp/EphyraSoftware:test-a?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=NL6ZHWZXRNCNNIHQKDXK2Q4GGA3PKQD3'
      )

    await wrapper.get('button').trigger('submit')

    await flushPromises()

    // Check that the created code was put in the store
    const groupsStore = useGroupsStore()
    expect(groupsStore.groups).toHaveLength(1)
    const codes = groupsStore.groupCodes(groupId)
    expect(codes).toHaveLength(1)

    // Check that the created event was emitted
    expect('created' in wrapper.emitted()).toBe(true)
    const created = wrapper.emitted('created')
    expect(created).toHaveLength(1)

    // The emit event should have the same data as the store
    expect(created![0]).toEqual(codes)
  })

  it('handle server error', async () => {
    const wrapper = mount(CreateCode, {
      global: {
        provide: {
          client
        }
      },
      props: {
        groupId
      }
    })
    await wrapper
      .get('input[data-test-id="code-original"]')
      .setValue(
        'otpauth://totp/EphyraSoftware:test-a?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=NL6ZHWZXRNCNNIHQKDXK2Q4GGA3PKQD3'
      )

    setNextIsHttpErr()
    await wrapper.get('button').trigger('submit')

    await flushPromises()

    expect(wrapper.html()).toContain('Error creating your code: A test error')
  })

  it('create code manual', async () => {
    const wrapper = mount(CreateCode, {
      global: {
        provide: {
          client
        }
      },
      props: {
        groupId
      }
    })

    // Switch to the manual tab
    await wrapper.get('a[data-test-id="manual"]').trigger('click')

    // Fill out the form, leaving defaults in place
    await wrapper.get('input[data-test-id="code-provider"]').setValue('EphyraSoftware')

    await wrapper.get('input[data-test-id="code-name"]').setValue('test-a')

    await wrapper
      .get('input[data-test-id="code-secret"]')
      .setValue('NL6ZHWZXRNCNNIHQKDXK2Q4GGA3PKQD3')

    // Submit the form to creat a code
    await wrapper.get('button').trigger('submit')

    await flushPromises()

    // Check that the created code was put in the store
    const groupsStore = useGroupsStore()
    expect(groupsStore.groups).toHaveLength(1)
    const codes = groupsStore.groupCodes(groupId)
    expect(codes).toHaveLength(1)
  })
})
