import { describe, it, expect, beforeEach, beforeAll, afterAll } from 'vitest'
import CreateCodeGroup from '@/components/CreateCodeGroup.vue'
import { flushPromises, mount } from '@vue/test-utils'
import type { Pinia } from 'pinia'
import { createPinia, setActivePinia } from 'pinia'
import { useGroupsStore } from '@/stores/groups'
import axios from 'axios'
import type { AxiosInstance } from 'axios'
import {
  resetHttpMockServer,
  setNextIsHttpErr,
  startHttpMockServer,
  stopHttpMockServer
} from '@/support/httpMock'

describe('CreateCodeGroup', () => {
  let client: AxiosInstance
  let pinia: Pinia

  beforeAll(() => {
    startHttpMockServer()
  })

  afterAll(() => {
    stopHttpMockServer()
  })

  beforeEach(() => {
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
  })

  it('render', () => {
    const wrapper = mount(CreateCodeGroup, {
      global: {
        provide: {
          client
        }
      }
    })
    expect(wrapper.html()).toContain('Create a new group')
  })

  it('create a new group', async () => {
    const wrapper = mount(CreateCodeGroup, {
      global: {
        provide: {
          client
        }
      }
    })
    await wrapper.get('input').setValue('Test Group')

    await wrapper.get('button').trigger('submit')

    await flushPromises()

    // Check that the store has the new group
    const groupsStore = useGroupsStore()
    expect(groupsStore.groups).toHaveLength(1)
    const group = groupsStore.groups[0]

    // Check that the created event was emitted
    expect('created' in wrapper.emitted()).toBe(true)
    const created = wrapper.emitted('created')
    expect(created).toHaveLength(1)

    // The emit event should have the same data as the store
    expect(created![0]).toEqual([group])
  })

  it('handle server error', async () => {
    const wrapper = mount(CreateCodeGroup, {
      global: {
        provide: {
          client
        }
      }
    })
    await wrapper.get('input').setValue('Test Group')

    setNextIsHttpErr()
    await wrapper.get('button').trigger('submit')

    await flushPromises()

    expect(wrapper.html()).toContain('Error creating your group: A test error')
  })
})
