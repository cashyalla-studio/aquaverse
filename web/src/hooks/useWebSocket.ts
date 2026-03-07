import { useEffect, useRef, useState, useCallback } from 'react'

export type WsStatus = 'connecting' | 'open' | 'closed' | 'error'

interface UseWebSocketOptions {
  onMessage?: (data: unknown) => void
  reconnectDelay?: number
  maxReconnectAttempts?: number
}

export function useWebSocket(url: string | null, options: UseWebSocketOptions = {}) {
  const { onMessage, reconnectDelay = 3000, maxReconnectAttempts = 5 } = options
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectAttempts = useRef(0)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>()
  const [status, setStatus] = useState<WsStatus>('closed')
  const mountedRef = useRef(true)

  const connect = useCallback(() => {
    if (!url || !mountedRef.current) return
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    setStatus('connecting')
    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      if (!mountedRef.current) return
      setStatus('open')
      reconnectAttempts.current = 0
    }

    ws.onmessage = (event) => {
      if (!mountedRef.current) return
      try {
        const data = JSON.parse(event.data as string)
        onMessage?.(data)
      } catch {
        // ignore parse errors
      }
    }

    ws.onerror = () => {
      if (!mountedRef.current) return
      setStatus('error')
    }

    ws.onclose = () => {
      if (!mountedRef.current) return
      setStatus('closed')
      // 재연결 (지수 백오프)
      if (reconnectAttempts.current < maxReconnectAttempts) {
        const delay = reconnectDelay * Math.pow(1.5, reconnectAttempts.current)
        reconnectAttempts.current++
        reconnectTimer.current = setTimeout(connect, delay)
      }
    }
  }, [url, onMessage, reconnectDelay, maxReconnectAttempts])

  const send = useCallback((data: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data))
    }
  }, [])

  const disconnect = useCallback(() => {
    clearTimeout(reconnectTimer.current)
    reconnectAttempts.current = maxReconnectAttempts // 재연결 방지
    wsRef.current?.close()
  }, [maxReconnectAttempts])

  useEffect(() => {
    mountedRef.current = true
    connect()
    return () => {
      mountedRef.current = false
      clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  return { status, send, disconnect }
}
