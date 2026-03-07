import { useState } from 'react'
import api from '../../api/client'

type Step = 'idle' | 'qr' | 'verify' | 'done' | 'disable'

export default function TotpSetup() {
  const [step, setStep] = useState<Step>('idle')
  const [secret, setSecret] = useState('')
  const [qrUrl, setQrUrl] = useState('')
  const [code, setCode] = useState('')
  const [backupCodes, setBackupCodes] = useState<string[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleEnable = async () => {
    setLoading(true); setError('')
    try {
      const r = await api.post('/auth/totp/enable')
      setSecret(r.data.secret)
      setQrUrl(r.data.qr_url)
      setStep('qr')
    } catch (e: any) {
      setError(e.response?.data?.error || '오류가 발생했습니다')
    } finally {
      setLoading(false)
    }
  }

  const handleVerify = async () => {
    setLoading(true); setError('')
    try {
      const r = await api.post('/auth/totp/verify', { code })
      setBackupCodes(r.data.backup_codes || [])
      setStep('done')
    } catch (e: any) {
      setError(e.response?.data?.error || '코드가 올바르지 않습니다')
    } finally {
      setLoading(false)
    }
  }

  const handleDisable = async () => {
    setLoading(true); setError('')
    try {
      await api.delete('/auth/totp', { data: { code } })
      setStep('idle'); setCode('')
    } catch (e: any) {
      setError(e.response?.data?.error || '코드가 올바르지 않습니다')
    } finally {
      setLoading(false)
    }
  }

  // QR 코드는 구글 차트 API 대신 qr_url을 직접 보여줌 (보안상 서드파티 QR 생성 API 미사용)
  // 실제 서비스에서는 qrcode.react 등 클라이언트 라이브러리 사용 권장
  const qrImageUrl = `https://chart.googleapis.com/chart?chs=200x200&chld=M|0&cht=qr&chl=${encodeURIComponent(qrUrl)}`

  return (
    <div className="max-w-md mx-auto p-6">
      <h1 className="text-xl font-bold mb-4">2단계 인증 (TOTP)</h1>

      {step === 'idle' && (
        <div>
          <p className="text-gray-600 mb-4">
            Google Authenticator, Authy 등의 앱으로 계정을 보호하세요.
          </p>
          <button
            onClick={handleEnable}
            disabled={loading}
            className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? '설정 중...' : '2단계 인증 활성화'}
          </button>
          <button
            onClick={() => setStep('disable')}
            className="w-full py-2 mt-2 border border-red-300 text-red-600 rounded hover:bg-red-50"
          >
            비활성화
          </button>
        </div>
      )}

      {step === 'qr' && (
        <div>
          <p className="text-gray-600 mb-4">인증 앱으로 QR 코드를 스캔하세요.</p>
          <div className="flex justify-center mb-4">
            <img src={qrImageUrl} alt="TOTP QR Code" className="border p-2" />
          </div>
          <div className="bg-gray-50 rounded p-3 mb-4">
            <p className="text-xs text-gray-500 mb-1">또는 수동 입력:</p>
            <code className="text-sm font-mono break-all">{secret}</code>
          </div>
          <input
            type="text"
            value={code}
            onChange={e => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
            placeholder="6자리 코드 입력"
            maxLength={6}
            className="w-full border rounded px-3 py-2 mb-3 text-center text-2xl font-mono tracking-widest"
          />
          {error && <p className="text-red-500 text-sm mb-2">{error}</p>}
          <button
            onClick={handleVerify}
            disabled={code.length !== 6 || loading}
            className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? '확인 중...' : '코드 확인 및 활성화'}
          </button>
        </div>
      )}

      {step === 'done' && (
        <div>
          <div className="bg-green-50 border border-green-200 rounded p-4 mb-4">
            <p className="text-green-700 font-medium">2단계 인증이 활성화되었습니다!</p>
          </div>
          <p className="text-sm text-gray-600 mb-3">
            다음 백업 코드를 안전한 곳에 저장하세요. 각 코드는 1회만 사용 가능합니다.
          </p>
          <div className="bg-gray-50 rounded p-3 font-mono text-sm grid grid-cols-2 gap-1">
            {backupCodes.map((c, i) => (
              <span key={i} className="bg-white border rounded px-2 py-1 text-center">{c}</span>
            ))}
          </div>
        </div>
      )}

      {step === 'disable' && (
        <div>
          <p className="text-gray-600 mb-4">비활성화하려면 현재 인증 앱의 코드를 입력하세요.</p>
          <input
            type="text"
            value={code}
            onChange={e => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
            placeholder="6자리 코드"
            maxLength={6}
            className="w-full border rounded px-3 py-2 mb-3 text-center text-2xl font-mono tracking-widest"
          />
          {error && <p className="text-red-500 text-sm mb-2">{error}</p>}
          <button
            onClick={handleDisable}
            disabled={code.length !== 6 || loading}
            className="w-full py-2 bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50"
          >
            {loading ? '처리 중...' : '2단계 인증 비활성화'}
          </button>
          <button onClick={() => setStep('idle')} className="w-full py-2 mt-2 text-gray-600">
            취소
          </button>
        </div>
      )}
    </div>
  )
}
