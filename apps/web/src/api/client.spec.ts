import { beforeEach, describe, expect, it, vi } from 'vitest'
import { apiRequest } from './client'

describe('apiRequest', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('adds JSON content type for serialized request bodies', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
    )

    await apiRequest('/matches', { method: 'POST', body: JSON.stringify({ difficulty: 1 }) })

    const [, init] = fetchMock.mock.calls[0]!
    expect(new Headers(init?.headers).get('Content-Type')).toBe('application/json')
  })

  it('lets the browser set multipart headers for FormData', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
    )
    const form = new FormData()
    form.append('file', new Blob(['a3a4']), 'demo.iccs')

    await apiRequest('/records/imports', { method: 'POST', body: form })

    const [, init] = fetchMock.mock.calls[0]!
    expect(new Headers(init?.headers).has('Content-Type')).toBe(false)
  })
})
