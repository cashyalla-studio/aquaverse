import { useState, useEffect } from 'react'
import api from '../api/client'

interface CitesCheckResult {
  is_blocked: boolean
  has_warning: boolean
  appendix?: 'I' | 'II' | 'III'
  message?: string
  is_invasive_kr: boolean
}

interface CitesWarningProps {
  scientificName: string | undefined
}

export function CitesWarning({ scientificName }: CitesWarningProps) {
  const [result, setResult] = useState<CitesCheckResult | null>(null)

  useEffect(() => {
    if (!scientificName || scientificName.trim().length < 3) {
      setResult(null)
      return
    }
    const timer = setTimeout(async () => {
      try {
        const res = await api.get<CitesCheckResult>(`/cites/check`, {
          params: { scientific_name: scientificName },
        })
        setResult(res.data)
      } catch {
        setResult(null)
      }
    }, 600) // 600ms 디바운스
    return () => clearTimeout(timer)
  }, [scientificName])

  if (!result?.has_warning) return null

  return (
    <div className={`rounded-xl p-4 flex gap-3 ${
      result.is_blocked
        ? 'bg-red-50 border border-red-200'
        : 'bg-amber-50 border border-amber-200'
    }`}>
      <span className="text-2xl flex-shrink-0">
        {result.is_blocked ? '🚫' : '⚠️'}
      </span>
      <div>
        <p className={`font-semibold text-sm ${
          result.is_blocked ? 'text-red-800' : 'text-amber-800'
        }`}>
          {result.is_blocked
            ? 'CITES 보호종 — 거래 차단'
            : `CITES 부속서 ${result.appendix} 보호종 — 주의 필요`}
        </p>
        <p className={`text-xs mt-1 ${
          result.is_blocked ? 'text-red-600' : 'text-amber-600'
        }`}>
          {result.message}
        </p>
        {result.is_invasive_kr && (
          <p className="text-xs mt-1 text-orange-600 font-medium">
            🇰🇷 한국 생태계 교란종 — 자연 방류 시 불법
          </p>
        )}
      </div>
    </div>
  )
}
