import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useWebSocket } from '../../hooks/useWebSocket'
import { useAuthStore } from '../../store/authStore'

interface ChatMessage {
  message_id: number
  room_id: number
  sender_id: string
  content: string
  msg_type: 'TEXT' | 'IMAGE' | 'SYSTEM'
  created_at: string
}

interface WsPayload {
  type: 'message' | 'history' | 'error' | 'system'
  room_id: number
  messages?: ChatMessage[]
  sender_id?: string
  content?: string
  msg_type?: string
  message_id?: number
  created_at?: string
  error?: string
}

export default function TradeChat() {
  const { tradeId } = useParams<{ tradeId: string }>()
  const navigate = useNavigate()
  const { user, accessToken } = useAuthStore()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [isConnecting, setIsConnecting] = useState(true)
  const bottomRef = useRef<HTMLDivElement>(null)

  // API 주소에서 WS 주소 생성
  const apiBase = import.meta.env.VITE_API_URL || 'http://localhost:8080'
  const wsBase = apiBase.replace(/^http/, 'ws')
  const wsUrl =
    tradeId && accessToken
      ? `${wsBase}/api/v1/trades/${tradeId}/chat?token=${accessToken}`
      : null

  const handleMessage = useCallback((data: unknown) => {
    const payload = data as WsPayload
    if (payload.type === 'history' && payload.messages) {
      setMessages(payload.messages)
      setIsConnecting(false)
    } else if (payload.type === 'message') {
      const msg: ChatMessage = {
        message_id: payload.message_id!,
        room_id: payload.room_id,
        sender_id: payload.sender_id!,
        content: payload.content!,
        msg_type: (payload.msg_type as 'TEXT') || 'TEXT',
        created_at: payload.created_at!,
      }
      setMessages((prev) => [...prev, msg])
    }
  }, [])

  const { status, send } = useWebSocket(wsUrl, { onMessage: handleMessage })

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSend = () => {
    const content = input.trim()
    if (!content || status !== 'open') return
    send({ content })
    setInput('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const statusColor: Record<typeof status, string> = {
    open: 'bg-green-500',
    connecting: 'bg-yellow-500',
    closed: 'bg-gray-400',
    error: 'bg-red-500',
  }

  const statusText: Record<typeof status, string> = {
    open: '연결됨',
    connecting: '연결 중...',
    closed: '연결 끊김',
    error: '오류',
  }

  return (
    <div className="flex flex-col h-screen max-w-2xl mx-auto bg-white">
      {/* 헤더 */}
      <div className="flex items-center gap-3 px-4 py-3 border-b bg-white sticky top-0 z-10">
        <button onClick={() => navigate(-1)} className="text-gray-500 hover:text-gray-800">
          ←
        </button>
        <h1 className="flex-1 font-semibold text-lg">거래 채팅</h1>
        <div className="flex items-center gap-1.5 text-sm text-gray-500">
          <span className={`w-2 h-2 rounded-full ${statusColor[status]}`} />
          {statusText[status]}
        </div>
      </div>

      {/* 메시지 목록 */}
      <div className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
        {isConnecting && (
          <div className="text-center text-gray-400 py-8">채팅 기록 불러오는 중...</div>
        )}
        {messages.map((msg) => {
          const isMine = msg.sender_id === user?.id
          return (
            <div
              key={msg.message_id}
              className={`flex ${isMine ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-xs lg:max-w-md px-4 py-2 rounded-2xl text-sm ${
                  isMine
                    ? 'bg-blue-500 text-white rounded-br-sm'
                    : 'bg-gray-100 text-gray-900 rounded-bl-sm'
                }`}
              >
                <p className="whitespace-pre-wrap break-words">{msg.content}</p>
                <p className={`text-xs mt-1 ${isMine ? 'text-blue-100' : 'text-gray-400'}`}>
                  {new Date(msg.created_at).toLocaleTimeString('ko-KR', {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </p>
              </div>
            </div>
          )
        })}
        <div ref={bottomRef} />
      </div>

      {/* 입력창 */}
      <div className="px-4 py-3 border-t bg-white">
        <div className="flex gap-2 items-end">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={status === 'open' ? '메시지를 입력하세요...' : '연결 중...'}
            disabled={status !== 'open'}
            rows={1}
            className="flex-1 resize-none rounded-xl border border-gray-300 px-3 py-2 text-sm
                       focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50
                       max-h-32 overflow-y-auto"
            style={{ minHeight: '40px' }}
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || status !== 'open'}
            className="px-4 py-2 bg-blue-500 text-white rounded-xl text-sm font-medium
                       hover:bg-blue-600 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            전송
          </button>
        </div>
        <p className="text-xs text-gray-400 mt-1">Enter로 전송, Shift+Enter로 줄바꿈</p>
      </div>
    </div>
  )
}
