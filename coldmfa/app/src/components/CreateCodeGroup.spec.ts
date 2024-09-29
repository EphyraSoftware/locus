import { describe, it, expect, beforeEach } from 'vitest'
import CreateCodeGroup from '@/components/CreateCodeGroup.vue'
import { flushPromises, mount } from '@vue/test-utils'
import type { Pinia } from 'pinia'
import { createPinia, setActivePinia } from 'pinia'
import { useGroupsStore } from '@/stores/groups'
import axios from 'axios'
import type { AxiosInstance } from 'axios'
import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { nanoid } from 'nanoid'

let nextIsHttpErr = false

const handlers = [
  http.options('*', () => {
    return new Response(null, {
      status: 200,
      headers: {
        Allow: 'GET,HEAD,POST'
      }
    })
  }),

  http.post('http://127.0.0.1:3000/coldmfa/api/groups', () => {
    if (nextIsHttpErr) {
      nextIsHttpErr = false
      return HttpResponse.json({ error: 'A test error' }, { status: 500 })
    }

    return HttpResponse.json(
      {
        id: nanoid(),
        name: 'Test Group',
        codes: []
      },
      { status: 201 }
    )
  })
]

export const server = setupServer(...handlers)

server.listen()

describe('CreateCodeGroup', () => {
  let client: AxiosInstance
  let pinia: Pinia

  beforeEach(() => {
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

    const groupsStore = useGroupsStore()
    expect(groupsStore.groups.length).toBe(1)
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

    nextIsHttpErr = true
    await wrapper.get('button').trigger('submit')

    await flushPromises()

    expect(wrapper.html()).toContain('Error creating your group: A test error')
  })
})
