import { afterAll, beforeAll, beforeEach, describe, it, expect } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import CodeList from '@/components/CodeList.vue'
import type { AxiosInstance, AxiosResponse } from 'axios'
import type { Pinia } from 'pinia'
import { resetHttpMockServer, startHttpMockServer, stopHttpMockServer } from '@/support/httpMock'
import axios from 'axios'
import { createPinia, setActivePinia } from 'pinia'
import type { CodeGroup, CodeSummary } from '@/types'
import { useGroupsStore } from '@/stores/groups'

describe('CodeList', () => {
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

    const codeResp1: AxiosResponse<CodeSummary> = await client.post(
      `http://127.0.0.1:3000/coldmfa/api/groups/${groupId}/codes`,
      {
        original:
          'otpauth://totp/EphyraSoftware:test-a?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=NL6ZHWZXRNCNNIHQKDXK2Q4GGA3PKQD3'
      }
    )
    groupsStore.addCodeToGroup(groupId, codeResp1.data)

    const codeResp2: AxiosResponse<CodeSummary> = await client.post(
      `http://127.0.0.1:3000/coldmfa/api/groups/${groupId}/codes`,
      {
        original:
          'otpauth://totp/EphyraSoftware:test-b?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=A23DJ4WDRR2XFPDKBUQ5ZLZN6KVIIIC4'
      }
    )
    groupsStore.addCodeToGroup(groupId, codeResp2.data)
  })

  it('renders', () => {
    const wrapper = mount(CodeList, {
      props: {
        groupId,
        showUpdateNameButton: false
      },
      global: {
        provide: {
          client
        }
      }
    })

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')
    expect(wrapper.html()).toContain('EphyraSoftware:test-b')
  })

  it('delete codes', async () => {
    const wrapper = mount(CodeList, {
      props: {
        groupId,
        showUpdateNameButton: false
      },
      global: {
        provide: {
          client
        }
      }
    })

    const deleteButtons = wrapper.findAll("button[data-test-id='delete']")
    expect(deleteButtons.length).toBe(2)
    for (const deleteButton of deleteButtons) {
      for (let i = 0; i < 5; i++) {
        await deleteButton.trigger('click')
      }
    }
    await flushPromises()

    expect(wrapper.html()).toContain('No codes yet')
  })

  it('export one at a time', async () => {
    const wrapper = mount(CodeList, {
      props: {
        groupId,
        showUpdateNameButton: false
      },
      global: {
        provide: {
          client
        }
      }
    })

    const exportButtons = wrapper.findAll("button[data-test-id='export']")
    expect(exportButtons.length).toBe(2)
    for (const exportButton of exportButtons) {
      await exportButton.trigger('click')
      await flushPromises()

      const closeExportButtons = wrapper.findAll("button[data-test-id='close-export']")
      expect(closeExportButtons.length).toBe(1)

      // Rough check that the QR code is visible
      expect(wrapper.get('img').isVisible()).toBe(true)
      expect(wrapper.get('img')).not.toContain('QR code')
    }

    let closeExportButtons = wrapper.findAll("button[data-test-id='close-export']")
    expect(closeExportButtons.length).toBe(1)
    await closeExportButtons[0].trigger('click')

    closeExportButtons = wrapper.findAll("button[data-test-id='close-export']")
    expect(closeExportButtons.length).toBe(0)
  })

  it('show deleted', async () => {
    const wrapper = mount(CodeList, {
      props: {
        groupId,
        showUpdateNameButton: false
      },
      global: {
        provide: {
          client
        }
      }
    })

    const deleteButtons = wrapper.findAll("button[data-test-id='delete']")
    expect(deleteButtons.length).toBe(2)
    for (let i = 0; i < 5; i++) {
      await deleteButtons[0].trigger('click')
    }
    await flushPromises()

    expect(wrapper.html()).not.toContain('EphyraSoftware:test-a')
    expect(wrapper.html()).toContain('EphyraSoftware:test-b')

    await wrapper.get("input[data-test-id='show-deleted']").setValue(true)

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')
    expect(wrapper.html()).toContain('EphyraSoftware:test-b')

    expect(wrapper.html()).toContain('This code has been deleted')
  })

  it('sort', async () => {
    const wrapper = mount(CodeList, {
      props: {
        groupId,
        showUpdateNameButton: false
      },
      global: {
        provide: {
          client
        }
      }
    })

    // Change the created date of the second code to be newer
    // This is because the setup can create both codes in the same millisecond
    const groupsStore = useGroupsStore()

    const createdAt = groupsStore.groups[0]?.codes![1]?.createdAt
    if (createdAt) {
      groupsStore.groups[0]!.codes![1]!.createdAt = createdAt + 1000
    }

    expect(wrapper.html()).toContain('EphyraSoftware:test-a')
    expect(wrapper.html()).toContain('EphyraSoftware:test-b')

    const codeNames = wrapper.findAll("p[data-test-id='code-name']")
    expect(codeNames.length).toBe(2)
    expect(codeNames[0].text()).toBe('EphyraSoftware:test-a')
    expect(codeNames[1].text()).toBe('EphyraSoftware:test-b')

    await wrapper.get("select[data-test-id='sort-by']").setValue('create')

    const codeNamesByCreatedDate = wrapper.findAll("p[data-test-id='code-name']")
    expect(codeNamesByCreatedDate.length).toBe(2)
    expect(codeNamesByCreatedDate[0].text()).toBe('EphyraSoftware:test-b')
    expect(codeNamesByCreatedDate[1].text()).toBe('EphyraSoftware:test-a')

    await wrapper.get("select[data-test-id='sort-by']").setValue('alpha')

    const codeNamesByAlpha = wrapper.findAll("p[data-test-id='code-name']")
    expect(codeNamesByAlpha.length).toBe(2)
    expect(codeNamesByAlpha[0].text()).toBe('EphyraSoftware:test-a')
    expect(codeNamesByAlpha[1].text()).toBe('EphyraSoftware:test-b')
  })

  it('sort by preferred name', async () => {
    const wrapper = mount(CodeList, {
      props: {
        groupId,
        showUpdateNameButton: true
      },
      global: {
        provide: {
          client
        }
      }
    })

    const renameButtons = wrapper.findAll("button[data-test-id='rename']")
    expect(renameButtons.length).toBe(2)

    const codeNames = wrapper.findAll("p[data-test-id='code-name']")
    expect(codeNames.length).toBe(2)
    expect(codeNames[1].text()).toBe('EphyraSoftware:test-b')
    codeNames[1].element.textContent = 'aaa'

    await renameButtons[1].trigger('click')
    await flushPromises()

    const updatedCodeNames = wrapper.findAll("p[data-test-id='code-name']")
    expect(updatedCodeNames.length).toBe(2)
    expect(updatedCodeNames[0].text()).toBe('aaa')
    expect(updatedCodeNames[1].text()).toBe('EphyraSoftware:test-a')
  })
})
