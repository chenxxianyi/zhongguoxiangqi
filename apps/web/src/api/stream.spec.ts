import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { connectMatchStream } from './stream'

class MockWebSocket {
  static instances: MockWebSocket[] = []

  onopen: (() => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  onclose: (() => void) | null = null
  onerror: (() => void) | null = null

  constructor(public url: string) {
    MockWebSocket.instances.push(this)
  }

  close() {}
}

describe('match stream cursor', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    MockWebSocket.instances = []
    vi.stubGlobal('WebSocket', MockWebSocket)
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('reconnects after the last successfully consumed event', async () => {
    const stream = connectMatchStream('match-1')
    const handler = vi.fn()
    stream.onEvent(handler)

    const first = MockWebSocket.instances[0]!
    first.onmessage?.({
      data: JSON.stringify({
        eventId: 'event-1',
        matchId: 'match-1',
        matchVersion: 2,
        type: 'match.snapshot',
        timestamp: '2026-07-16T00:00:00Z',
        payload: {},
      }),
    } as MessageEvent)
    first.onclose?.()

    await vi.advanceTimersByTimeAsync(1000)

    expect(handler).toHaveBeenCalledTimes(1)
    expect(MockWebSocket.instances[1]?.url).toContain('afterEventId=event-1')
    stream.close()
  })
})
