import { useState, useRef, useEffect, useCallback } from 'react'
import { MessageCircle, X, Send, Bot, Loader2 } from 'lucide-react'
import { clsx } from 'clsx'

const API_BASE = import.meta.env.VITE_API_URL ?? ''
const SESSION_STORAGE_KEY = 'finara_chat_session_id'

interface Message {
  role: 'user' | 'assistant'
  content: string
}

interface Source {
  id: number
  name: string
}

interface AskResponse {
  answer: string
  session_id: string
  sources?: Source[]
}

export default function ChatBot() {
  const [open, setOpen] = useState(false)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [sessionID, setSessionID] = useState<string>(() =>
    localStorage.getItem(SESSION_STORAGE_KEY) ?? '',
  )
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // 패널이 열릴 때 입력창에 포커스
  useEffect(() => {
    if (open) {
      setTimeout(() => inputRef.current?.focus(), 120)
    }
  }, [open])

  // 새 메시지가 추가될 때 스크롤 하단으로 이동
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const sendMessage = useCallback(async () => {
    const question = input.trim()
    if (!question || loading) return

    const userMsg: Message = { role: 'user', content: question }
    setMessages((prev) => [...prev, userMsg])
    setInput('')
    setLoading(true)

    try {
      const token = localStorage.getItem('access_token')
      const headers: Record<string, string> = { 'Content-Type': 'application/json' }
      if (token) headers['Authorization'] = `Bearer ${token}`

      const res = await fetch(`${API_BASE}/api/v1/chat/ask`, {
        method: 'POST',
        headers,
        body: JSON.stringify({ question, session_id: sessionID }),
      })

      if (!res.ok) throw new Error(`서버 오류 (${res.status})`)

      const data: AskResponse = await res.json()

      // 세션 ID 저장 (최초 응답 시 세팅)
      if (data.session_id && data.session_id !== sessionID) {
        setSessionID(data.session_id)
        localStorage.setItem(SESSION_STORAGE_KEY, data.session_id)
      }

      const assistantContent = data.sources && data.sources.length > 0
        ? `${data.answer}\n\n_참고 어종: ${data.sources.map((s) => s.name).join(', ')}_`
        : data.answer

      setMessages((prev) => [...prev, { role: 'assistant', content: assistantContent }])
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        {
          role: 'assistant',
          content: '죄송합니다. 일시적인 오류가 발생했습니다. 잠시 후 다시 시도해 주세요.',
        },
      ])
    } finally {
      setLoading(false)
    }
  }, [input, loading, sessionID])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  return (
    <>
      {/* ── 채팅 패널 ── */}
      {open && (
        <div
          className={clsx(
            'fixed bottom-20 right-4 z-50',
            'w-80 h-96 flex flex-col',
            'bg-white rounded-2xl shadow-2xl border border-gray-100',
            'overflow-hidden',
          )}
        >
          {/* 헤더 */}
          <div className="flex items-center justify-between px-4 py-3 bg-gradient-to-r from-primary-500 to-cyan-500 text-white flex-shrink-0">
            <div className="flex items-center gap-2">
              <Bot size={18} />
              <span className="font-semibold text-sm">Finara AI 어시스턴트</span>
            </div>
            <button
              onClick={() => setOpen(false)}
              className="p-1 rounded-lg hover:bg-white/20 transition-colors"
              aria-label="채팅창 닫기"
            >
              <X size={16} />
            </button>
          </div>

          {/* 메시지 영역 */}
          <div className="flex-1 overflow-y-auto px-3 py-3 space-y-2 bg-gray-50">
            {messages.length === 0 && (
              <div className="text-center text-gray-400 text-xs mt-6 px-4">
                <Bot size={32} className="mx-auto mb-2 text-gray-300" />
                <p>열대어 사육에 대해 무엇이든 물어보세요!</p>
              </div>
            )}
            {messages.map((msg, idx) => (
              <div
                key={idx}
                className={clsx(
                  'flex',
                  msg.role === 'user' ? 'justify-end' : 'justify-start',
                )}
              >
                <div
                  className={clsx(
                    'max-w-[85%] px-3 py-2 rounded-2xl text-sm leading-relaxed whitespace-pre-wrap',
                    msg.role === 'user'
                      ? 'bg-primary-500 text-white rounded-br-sm'
                      : 'bg-white text-gray-800 shadow-sm border border-gray-100 rounded-bl-sm',
                  )}
                >
                  {msg.content}
                </div>
              </div>
            ))}
            {loading && (
              <div className="flex justify-start">
                <div className="bg-white border border-gray-100 shadow-sm rounded-2xl rounded-bl-sm px-3 py-2">
                  <Loader2 size={16} className="animate-spin text-primary-500" />
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* 입력 영역 */}
          <div className="flex items-center gap-2 px-3 py-2.5 border-t border-gray-100 bg-white flex-shrink-0">
            <input
              ref={inputRef}
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="질문을 입력하세요..."
              disabled={loading}
              className={clsx(
                'flex-1 text-sm bg-gray-50 border border-gray-200 rounded-xl px-3 py-2',
                'focus:outline-none focus:ring-2 focus:ring-primary-300 focus:border-transparent',
                'disabled:opacity-50',
              )}
            />
            <button
              onClick={sendMessage}
              disabled={loading || !input.trim()}
              className={clsx(
                'p-2 rounded-xl transition-colors flex-shrink-0',
                loading || !input.trim()
                  ? 'text-gray-300 bg-gray-100 cursor-not-allowed'
                  : 'text-white bg-primary-500 hover:bg-primary-600',
              )}
              aria-label="전송"
            >
              <Send size={16} />
            </button>
          </div>
        </div>
      )}

      {/* ── FAB 버튼 ── */}
      <button
        onClick={() => setOpen((prev) => !prev)}
        className={clsx(
          'fixed bottom-20 right-4 z-50',
          'lg:bottom-6 lg:right-6',
          'w-12 h-12 rounded-full shadow-lg flex items-center justify-center',
          'bg-gradient-to-br from-primary-500 to-cyan-500 text-white',
          'hover:scale-105 active:scale-95 transition-transform',
          open && 'hidden',
        )}
        aria-label="AI 챗봇 열기"
      >
        <MessageCircle size={22} />
      </button>
    </>
  )
}
