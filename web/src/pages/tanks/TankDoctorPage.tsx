import { useState } from 'react'
import { useParams } from 'react-router-dom'
import api from '../../api/client'

interface FishState {
  fish_name: string
  status: 'good' | 'warning' | 'danger'
  issue: string
  suggestion: string
}

interface Diagnosis {
  tank_id: number
  summary: string
  fish_states: FishState[]
  actions: string[]
  created_at: string
}

const statusIcon: Record<string, string> = {
  good: '🟢',
  warning: '🟡',
  danger: '🔴',
}

const statusLabel: Record<string, string> = {
  good: '양호',
  warning: '주의',
  danger: '위험',
}

export default function TankDoctorPage() {
  const { id } = useParams<{ id: string }>()
  const [form, setForm] = useState({
    temp_c: '', ph: '', ammonia_ppm: '', nitrite_ppm: '', nitrate_ppm: '',
  })
  const [diagnosis, setDiagnosis] = useState<Diagnosis | null>(null)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const handleSave = async () => {
    setSaving(true)
    setError('')
    try {
      const body: Record<string, number | undefined> = {}
      if (form.temp_c) body.temp_c = parseFloat(form.temp_c)
      if (form.ph) body.ph = parseFloat(form.ph)
      if (form.ammonia_ppm) body.ammonia_ppm = parseFloat(form.ammonia_ppm)
      if (form.nitrite_ppm) body.nitrite_ppm = parseFloat(form.nitrite_ppm)
      if (form.nitrate_ppm) body.nitrate_ppm = parseFloat(form.nitrate_ppm)
      await api.post(`/tanks/${id}/water-params`, body)
    } catch (e: any) {
      setError(e.message)
    } finally {
      setSaving(false)
    }
  }

  const handleDiagnose = async () => {
    setLoading(true)
    setError('')
    setDiagnosis(null)
    try {
      await handleSave()
      const res = await api.get(`/tanks/${id}/diagnosis`)
      setDiagnosis(res.data)
    } catch (e: any) {
      setError(e.response?.data?.error || e.message)
    } finally {
      setLoading(false)
    }
  }

  const fields = [
    { key: 'temp_c', label: '수온 (°C)', placeholder: '예: 26.0' },
    { key: 'ph', label: 'pH', placeholder: '예: 7.2' },
    { key: 'ammonia_ppm', label: '암모니아 (ppm)', placeholder: '예: 0.0' },
    { key: 'nitrite_ppm', label: '아질산 (ppm)', placeholder: '예: 0.0' },
    { key: 'nitrate_ppm', label: '질산 (ppm)', placeholder: '예: 10.0' },
  ]

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-2">AI 수조 주치의</h1>
      <p className="text-gray-500 text-sm mb-6">수질 수치를 입력하면 Claude AI가 수조 상태를 진단합니다.</p>

      <div className="bg-white rounded-xl shadow-sm border p-6 mb-6">
        <h2 className="font-semibold mb-4">수질 파라미터 입력</h2>
        <div className="grid grid-cols-2 gap-4">
          {fields.map(({ key, label, placeholder }) => (
            <div key={key}>
              <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
              <input
                type="number"
                step="0.1"
                placeholder={placeholder}
                value={(form as any)[key]}
                onChange={(e) => setForm((f) => ({ ...f, [key]: e.target.value }))}
                className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          ))}
        </div>

        {error && <p className="text-red-500 text-sm mt-3">{error}</p>}

        <button
          onClick={handleDiagnose}
          disabled={loading || saving}
          className="mt-6 w-full bg-blue-600 text-white rounded-lg py-3 font-semibold hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {loading ? 'AI 진단 중...' : 'AI 진단 받기'}
        </button>
      </div>

      {diagnosis && (
        <div className="bg-white rounded-xl shadow-sm border p-6">
          <h2 className="font-semibold mb-2">진단 결과</h2>
          <p className="text-gray-700 mb-4 bg-blue-50 rounded-lg p-3 text-sm">{diagnosis.summary}</p>

          {diagnosis.fish_states && diagnosis.fish_states.length > 0 && (
            <div className="mb-4">
              <h3 className="text-sm font-semibold text-gray-600 mb-2">어종별 상태</h3>
              <div className="space-y-2">
                {diagnosis.fish_states.map((fs, i) => (
                  <div key={i} className="flex items-start gap-3 p-3 rounded-lg bg-gray-50">
                    <span className="text-xl">{statusIcon[fs.status] || '⚪'}</span>
                    <div>
                      <p className="font-medium text-sm">
                        {fs.fish_name}{' '}
                        <span className={`text-xs px-2 py-0.5 rounded-full ${
                          fs.status === 'good' ? 'bg-green-100 text-green-700' :
                          fs.status === 'warning' ? 'bg-yellow-100 text-yellow-700' :
                          'bg-red-100 text-red-700'
                        }`}>{statusLabel[fs.status] || fs.status}</span>
                      </p>
                      {fs.issue && <p className="text-xs text-gray-500 mt-0.5">{fs.issue}</p>}
                      {fs.suggestion && <p className="text-xs text-blue-600 mt-0.5">→ {fs.suggestion}</p>}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {diagnosis.actions && diagnosis.actions.length > 0 && (
            <div>
              <h3 className="text-sm font-semibold text-gray-600 mb-2">권장 조치</h3>
              <ul className="space-y-1">
                {diagnosis.actions.map((action, i) => (
                  <li key={i} className="text-sm flex items-start gap-2">
                    <span className="text-blue-500 font-bold mt-0.5">{i + 1}.</span>
                    <span>{action}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          <p className="text-xs text-gray-400 mt-4 text-right">
            진단 시각: {new Date(diagnosis.created_at).toLocaleString('ko-KR')}
          </p>
        </div>
      )}
    </div>
  )
}
