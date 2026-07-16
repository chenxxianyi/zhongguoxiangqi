import type { MatchEvent } from './contracts'

const wsBaseUrl = import.meta.env.VITE_WS_BASE_URL || '/api/v1'

interface StreamConnection {
  onEvent: (handler: (event: MatchEvent) => void) => void
  close: () => void
}

function getWSURL(path: string): string {
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
  // 开发环境下 Vite 代理 WebSocket
  return `${proto}://${window.location.host}${wsBaseUrl}${path}`
}

/**
 * 创建到对局事件流的 WebSocket 连接。
 * 自动重连（指数退避），收到正确 JSON 事件后调用 onEvent。
 */
export function connectMatchStream(matchId: string): StreamConnection {
  const handlers: Array<(event: MatchEvent) => void> = []
  let ws: WebSocket | null = null
  let closed = false
  let lastEventId: string | null = null
  let reconnectDelay = 1000
  let reconnectTimer: number | undefined

  function connect() {
    if (closed) return

    const cursor = lastEventId
      ? `?afterEventId=${encodeURIComponent(lastEventId)}`
      : ''
    ws = new WebSocket(
      getWSURL(`/matches/${encodeURIComponent(matchId)}/stream${cursor}`),
    )

    ws.onopen = () => {
      reconnectDelay = 1000 // 连接成功后重置退避
    }

    ws.onmessage = (msg: MessageEvent) => {
      try {
        const event = JSON.parse(msg.data) as MatchEvent
        // 验证事件基本结构
        if (event && event.eventId && event.type) {
          handlers.forEach((h) => h(event))
          lastEventId = event.eventId
        }
      } catch {
        // 忽略心跳等非 JSON 消息
      }
    }

    ws.onclose = () => {
      ws = null
      if (!closed) {
        reconnectTimer = window.setTimeout(connect, reconnectDelay)
        reconnectDelay = Math.min(reconnectDelay * 2, 10000)
      }
    }

    ws.onerror = () => {
      // onclose 会跟随 onerror，所以不需要额外处理
    }
  }

  connect()

  return {
    onEvent(handler: (event: MatchEvent) => void) {
      handlers.push(handler)
    },
    close() {
      closed = true
      if (reconnectTimer !== undefined) {
        clearTimeout(reconnectTimer)
        reconnectTimer = undefined
      }
      if (ws) {
        ws.onclose = null // 阻止重连
        ws.close()
        ws = null
      }
    },
  }
}
