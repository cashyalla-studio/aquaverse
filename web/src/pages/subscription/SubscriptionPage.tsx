import { useEffect, useState } from 'react'
import api from '../../api/client'
import { useAuthStore } from '../../store/authStore'

interface Plan {
  id: string; name: string; price_krw: number; features: string[]
}
interface Subscription {
  plan: string; status: string; expires_at?: string; billing_amount: number
}

export default function SubscriptionPage() {
  const { user } = useAuthStore()
  const [plans, setPlans] = useState<Plan[]>([])
  const [mySub, setMySub] = useState<Subscription | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    api.get('/subscriptions/plans').then(r => setPlans(r.data.plans || []))
    if (user) {
      api.get('/subscriptions/me').then(r => setMySub(r.data))
    }
  }, [user])

  const handleSubscribe = async (planId: string) => {
    if (!user) { alert('로그인이 필요합니다'); return }
    if (planId === 'FREE') return
    setLoading(true)
    try {
      await api.post('/subscriptions/subscribe', { trial: true })
      const r = await api.get('/subscriptions/me')
      setMySub(r.data)
      alert('PRO 구독이 활성화됐습니다! (1개월 무료 체험)')
    } catch (e: any) {
      alert(e.response?.data?.error || '구독 실패')
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = async () => {
    if (!confirm('구독을 취소하시겠습니까?')) return
    setLoading(true)
    try {
      await api.post('/subscriptions/cancel')
      const r = await api.get('/subscriptions/me')
      setMySub(r.data)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-2">구독 플랜</h1>
      {mySub && (
        <div className={`mb-6 p-4 rounded-xl border-2 ${mySub.plan === 'PRO' ? 'border-blue-400 bg-blue-50' : 'border-gray-200'}`}>
          <p className="font-semibold">현재 플랜: <span className={mySub.plan === 'PRO' ? 'text-blue-600' : ''}>{mySub.plan}</span></p>
          {mySub.expires_at && (
            <p className="text-sm text-gray-500 mt-0.5">
              {new Date(mySub.expires_at) > new Date() ? '만료일' : '만료됨'}: {new Date(mySub.expires_at).toLocaleDateString('ko-KR')}
            </p>
          )}
          {mySub.plan === 'PRO' && mySub.status === 'ACTIVE' && (
            <button onClick={handleCancel} disabled={loading}
              className="mt-2 text-sm text-red-500 hover:underline disabled:opacity-50">
              구독 취소
            </button>
          )}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {plans.map(plan => (
          <div key={plan.id}
            className={`bg-white rounded-xl shadow-sm border p-6 ${plan.id === 'PRO' ? 'border-blue-300 relative overflow-hidden' : ''}`}>
            {plan.id === 'PRO' && (
              <div className="absolute top-0 right-0 bg-blue-600 text-white text-xs px-3 py-1 rounded-bl-xl">추천</div>
            )}
            <h2 className="text-xl font-bold">{plan.name}</h2>
            <p className="text-3xl font-bold mt-2">
              {plan.price_krw === 0 ? '무료' : `₩${plan.price_krw.toLocaleString()}`}
              {plan.price_krw > 0 && <span className="text-base font-normal text-gray-500">/월</span>}
            </p>
            <ul className="mt-4 space-y-2">
              {plan.features.map((f, i) => (
                <li key={i} className="flex items-start gap-2 text-sm">
                  <span className="text-green-500 mt-0.5">✓</span>
                  <span>{f}</span>
                </li>
              ))}
            </ul>
            {plan.id === 'PRO' && mySub?.plan !== 'PRO' && (
              <button onClick={() => handleSubscribe(plan.id)} disabled={loading}
                className="mt-6 w-full bg-blue-600 text-white py-3 rounded-xl font-semibold hover:bg-blue-700 disabled:opacity-50">
                {loading ? '처리 중...' : '1개월 무료 체험 시작'}
              </button>
            )}
            {plan.id === 'PRO' && mySub?.plan === 'PRO' && (
              <div className="mt-6 text-center text-green-600 font-semibold">✓ 구독 중</div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
