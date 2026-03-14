import { ref, onMounted, onUnmounted } from 'vue'

export interface WSEvent {
  type: string
  data: any
}

const listeners = new Set<(event: WSEvent) => void>()
let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null

// Global reactive state
export const wsConnected = ref(false)

function setConnected(val: boolean) {
  wsConnected.value = val
}

function connect() {
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${protocol}//${location.host}/api/ws`

  ws = new WebSocket(url)

  ws.onopen = () => setConnected(true)

  ws.onmessage = (e) => {
    try {
      const event: WSEvent = JSON.parse(e.data)
      listeners.forEach(fn => fn(event))
    } catch {}
  }

  ws.onclose = () => {
    ws = null
    setConnected(false)
    reconnectTimer = setTimeout(connect, 2000)
  }

  ws.onerror = () => {
    ws?.close()
  }
}

export function useWebSocket(handler: (event: WSEvent) => void) {
  onMounted(() => {
    listeners.add(handler)
    if (!ws) connect()
  })

  onUnmounted(() => {
    listeners.delete(handler)
    if (listeners.size === 0 && ws) {
      if (reconnectTimer) clearTimeout(reconnectTimer)
      ws.close()
      ws = null
    }
  })

  return { connected: wsConnected }
}
