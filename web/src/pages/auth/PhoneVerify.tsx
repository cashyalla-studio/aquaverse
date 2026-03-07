import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../../api/client'

type Step = 'input' | 'code' | 'done'

export default function PhoneVerify() {
  const navigate = useNavigate()
  const [step, setStep] = useState<Step>('input')
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSendCode = async () => {
    if (!phone || phone.length < 10) {
      setError('올바른 전화번호를 입력하세요')
      return
    }
    setLoading(true)
    setError('')
    try {
      await api.post('/phone/send', { phone_number: phone })
      setStep('code')
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } }
      setError(err.response?.data?.message || '인증번호 발송에 실패했습니다')
    } finally {
      setLoading(false)
    }
  }

  const handleVerify = async () => {
    if (!code || code.length !== 6) {
      setError('6자리 인증번호를 입력하세요')
      return
    }
    setLoading(true)
    setError('')
    try {
      await api.post('/phone/verify', { phone_number: phone, code })
      setStep('done')
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } }
      setError(err.response?.data?.message || '인증에 실패했습니다')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-sm p-8">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">전화번호 인증</h1>
        <p className="text-gray-500 mb-8 text-sm">
          안전한 거래를 위해 전화번호 인증이 필요합니다
        </p>

        {step === 'done' ? (
          <div className="text-center py-8">
            <div className="text-5xl mb-4">✅</div>
            <p className="text-lg font-semibold text-gray-900">인증 완료!</p>
            <p className="text-gray-500 text-sm mt-2">이제 제한 없이 거래하실 수 있습니다</p>
            <button
              onClick={() => navigate('/')}
              className="mt-6 w-full py-3 bg-blue-600 text-white rounded-xl font-medium hover:bg-blue-700"
            >
              홈으로 이동
            </button>
          </div>
        ) : step === 'input' ? (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">전화번호</label>
              <input
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))}
                placeholder="01012345678"
                className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            {error && <p className="text-red-500 text-sm">{error}</p>}
            <button
              onClick={handleSendCode}
              disabled={loading}
              className="w-full py-3 bg-blue-600 text-white rounded-xl font-medium hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? '발송 중...' : '인증번호 받기'}
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              <span className="font-medium">{phone}</span>으로 인증번호를 발송했습니다
            </p>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">인증번호</label>
              <input
                type="text"
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                placeholder="6자리 입력"
                maxLength={6}
                className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-center text-2xl tracking-widest"
              />
            </div>
            {error && <p className="text-red-500 text-sm">{error}</p>}
            <button
              onClick={handleVerify}
              disabled={loading || code.length !== 6}
              className="w-full py-3 bg-blue-600 text-white rounded-xl font-medium hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? '확인 중...' : '인증 확인'}
            </button>
            <button
              onClick={() => {
                setStep('input')
                setCode('')
                setError('')
              }}
              className="w-full py-3 text-gray-500 text-sm hover:underline"
            >
              번호 다시 입력
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
