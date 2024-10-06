import { afterAll, beforeAll, beforeEach, describe, it, expect } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { AxiosInstance } from 'axios'
import type { Pinia } from 'pinia'
import { resetHttpMockServer, startHttpMockServer, stopHttpMockServer } from '@/support/httpMock'
import axios from 'axios'
import { createPinia, setActivePinia } from 'pinia'
import App from '@/App.vue'

describe('CodeGroups', () => {
  let client: AxiosInstance
  let pinia: Pinia

  const groupMsg = 'Create a new group'
  const backupMsg = 'Enter a password to encrypt the backup with'
  const restoreMsg = 'Enter a password to decrypt the backup'

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
  })

  it('render', async () => {
    const wrapper = mount(App, {
      props: {
        client
      }
    })

    await flushPromises()

    expect(wrapper.html()).toContain('Welcome, testuser')
    expect(wrapper.html()).toContain('Last backup at: Never')
  })

  it('show/hide new group', async () => {
    const wrapper = mount(App, {
      props: {
        client
      }
    })

    await flushPromises()

    await wrapper.get('button[data-test-id="new-group"]').trigger('click')

    expect(wrapper.html()).toContain(groupMsg)

    await wrapper.get('button[data-test-id="new-group"]').trigger('click')

    expect(wrapper.html()).not.toContain(groupMsg)
  })

  it('show/hide backup', async () => {
    const wrapper = mount(App, {
      props: {
        client
      }
    })

    await flushPromises()

    await wrapper.get('button[data-test-id="start-backup"]').trigger('click')

    expect(wrapper.html()).toContain(backupMsg)

    await wrapper.get('button[data-test-id="start-backup"]').trigger('click')

    expect(wrapper.html()).not.toContain(backupMsg)
  })

  it('show/hide restore', async () => {
    const wrapper = mount(App, {
      props: {
        client
      }
    })

    await flushPromises()

    await wrapper.get('button[data-test-id="start-restore"]').trigger('click')

    expect(wrapper.html()).toContain(restoreMsg)

    await wrapper.get('button[data-test-id="start-restore"]').trigger('click')

    expect(wrapper.html()).not.toContain(restoreMsg)
  })

  it('show/hide one at a time', async () => {
    const wrapper = mount(App, {
      props: {
        client
      }
    })

    await flushPromises()

    expect(wrapper.html()).not.toContain(groupMsg)
    expect(wrapper.html()).not.toContain(backupMsg)
    expect(wrapper.html()).not.toContain(restoreMsg)

    await wrapper.get('button[data-test-id="new-group"]').trigger('click')

    expect(wrapper.html()).toContain(groupMsg)
    expect(wrapper.html()).not.toContain(backupMsg)
    expect(wrapper.html()).not.toContain(restoreMsg)

    await wrapper.get('button[data-test-id="start-backup"]').trigger('click')

    expect(wrapper.html()).not.toContain(groupMsg)
    expect(wrapper.html()).toContain(backupMsg)
    expect(wrapper.html()).not.toContain(restoreMsg)

    await wrapper.get('button[data-test-id="start-restore"]').trigger('click')

    expect(wrapper.html()).not.toContain(groupMsg)
    expect(wrapper.html()).not.toContain(backupMsg)
    expect(wrapper.html()).toContain(restoreMsg)
  })
})
