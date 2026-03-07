import { useState, useEffect } from 'react'
import api from '../api/client'

interface EscrowPanelProps {
  tradeId: number
  isBuyer: boolean
  tradeAmount: number
  currency?: string
}

type EscrowStatus = 'PENDING' | 'FUNDED' | 'RELEASED' | 'REFUNDED' | 'DISPUTED'

export function EscrowPanel({ tradeId, isBuyer, tradeAmount, currency = 'KRW' }: EscrowPanelProps) {
  const [status, setStatus] = useState<EscrowStatus | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    api.get<{ escrow_status: EscrowStatus }>(`/trades/${tradeId}/escrow`)
      .then(res => setStatus(res.data.escrow_status))
      .catch(() => setStatus(null))
  }, [tradeId])

  const handleAction = async (action: 'fund' | 'release' | 'refund') => {
    setLoading(true)
    setError('')
    try {
      const res = await api.post<{ status: EscrowStatus }>(`/trades/${tradeId}/escrow/${action}`)
      setStatus(res.data.status as EscrowStatus)
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } }
      setError(err.response?.data?.message || '처리 중 오류가 발생했습니다')
    } finally {
      setLoading(false)
    }
  }

  if (status === null) return null

  const statusConfig: Record<EscrowStatus, { label: string; color: string; icon: string }> = {
    PENDING:  { label: '에스크로 대기',    color: 'text-gray-600 bg-gray-50 border-gray-200',   icon: '⏳' },
    FUNDED:   { label: '에스크로 보관 중', color: 'text-blue-700 bg-blue-50 border-blue-200',   icon: '🔒' },
    RELEASED: { label: '대금 지급 완료',   color: 'text-green-700 bg-green-50 border-green-200', icon: '✅' },
    REFUNDED: { label: '환불 완료',        color: 'text-orange-700 bg-orange-50 border-orange-200', icon: '↩️' },
    DISPUTED: { label: '분쟁 중',          color: 'text-red-700 bg-red-50 border-red-200',       icon: '⚠️' },
  }

  const cfg = statusConfig[status]
  const amountFormatted = new Intl.NumberFormat('ko-KR', {
    style: 'currency', currency,
  }).format(tradeAmount)

  return (
    <div className={`rounded-xl border p-4 ${cfg.color}`}>
      <div className="flex items-center gap-2 mb-3">
        <span className="text-xl">{cfg.icon}</span>
        <div>
          <p className="font-semibold text-sm">안전결제 (에스크로)</p>
          <p className="text-xs opacity-80">{cfg.label}</p>
        </div>
        <span className="ml-auto font-bold">{amountFormatted}</span>
      </div>

      <p className="text-xs opacity-70 mb-3">
        플랫폼이 대금을 안전하게 보관하고 거래 완료 후 판매자에게 지급합니다.
      </p>

      {error && <p className="text-red-600 text-xs mb-2">{error}</p>}

      <div className="flex gap-2">
        {isBuyer && status === 'PENDING' && (
          <button
            onClick={() => handleAction('fund')}
            disabled={loading}
            className="flex-1 py-2 bg-blue-600 text-white text-sm rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? '처리 중...' : '에스크로 입금'}
          </button>
        )}
        {isBuyer && status === 'FUNDED' && (
          <>
            <button
              onClick={() => handleAction('release')}
              disabled={loading}
              className="flex-1 py-2 bg-green-600 text-white text-sm rounded-lg font-medium hover:bg-green-700 disabled:opacity-50"
            >
              {loading ? '처리 중...' : '수령 확인 → 대금 지급'}
            </button>
            <button
              onClick={() => handleAction('refund')}
              disabled={loading}
              className="px-4 py-2 bg-red-100 text-red-700 text-sm rounded-lg font-medium hover:bg-red-200 disabled:opacity-50"
            >
              환불 요청
            </button>
          </>
        )}
      </div>
    </div>
  )
}
