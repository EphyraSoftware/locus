import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { nanoid } from 'nanoid'
import type { CodeGroup } from '@/types'

let nextIsHttpErr = false

let data: Record<string, CodeGroup> = {}

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

    const newGroup = {
      groupId: nanoid(),
      name: 'Test Group',
      codes: []
    }
    data[newGroup.groupId] = newGroup

    return HttpResponse.json(newGroup, { status: 201 })
  }),

  http.post(
    'http://127.0.0.1:3000/coldmfa/api/groups/:groupId/codes',
    async ({ request, params }) => {
      if (nextIsHttpErr) {
        nextIsHttpErr = false
        return HttpResponse.json({ error: 'A test error' }, { status: 500 })
      }

      const r = await request.json()
      if (!r || typeof r !== 'object') {
        throw new Error('expected create code request')
      }

      const newCode = {
        codeId: nanoid(),
        name: new URL(r['original']).href,
        deleted: false
      }

      const groupId = params['groupId'] as string
      const existingGroup = data[groupId]
      if (!existingGroup) {
        throw new Error('missing group')
      }
      if (!existingGroup.codes) {
        existingGroup.codes = []
      }

      existingGroup.codes.push(newCode)

      return HttpResponse.json(newCode, { status: 201 })
    }
  )
]

const server = setupServer(...handlers)

export const setNextIsHttpErr = () => {
  nextIsHttpErr = true
}

export const startHttpMockServer = () => {
  server.listen()
}

export const resetHttpMockServer = () => {
  nextIsHttpErr = false
  data = {}
  server.resetHandlers()
}

export const stopHttpMockServer = () => {
  server.close()
}
