import { useEffect, useRef, useCallback } from 'react'

type MessageHandler = (type: string, payload: unknown) => void

export function useWebSocket(token: string | null, onMessage: MessageHandler) {
  const wsRef = useRef<WebSocket | null>(null)

  const connect = useCallback(() => {
    if (!token) return
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const ws = new WebSocket(`${proto}://${location.host}/ws?token=${token}`)

    ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data) as { type: string; payload: unknown }
        onMessage(msg.type, msg.payload)
      } catch {
        // ignore malformed messages
      }
    }

    ws.onclose = () => {
      // Auto-reconnect after 3 seconds
      setTimeout(connect, 3000)
    }

    wsRef.current = ws
  }, [token, onMessage])

  useEffect(() => {
    connect()
    return () => {
      wsRef.current?.close()
    }
  }, [connect])
}
