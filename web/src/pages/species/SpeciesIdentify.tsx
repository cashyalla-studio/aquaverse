import { useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import api from '../../api/client'

interface Candidate {
  name: string
  scientific_name: string
  confidence: number
  description: string
  fish_data_id?: number
}

interface IdentifyResult {
  id: number
  candidates: Candidate[]
  processing_ms: number
}

export default function SpeciesIdentify() {
  const fileRef = useRef<HTMLInputElement>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [result, setResult] = useState<IdentifyResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleFile = (file: File) => {
    if (file.size > 5 * 1024 * 1024) {
      setError('이미지 크기는 5MB 이하여야 합니다.')
      return
    }
    const reader = new FileReader()
    reader.onload = e => setPreview(e.target?.result as string)
    reader.readAsDataURL(file)
    setError('')
    setResult(null)
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    const file = e.dataTransfer.files[0]
    if (file?.type.startsWith('image/')) handleFile(file)
  }

  const handleIdentify = async () => {
    if (!fileRef.current?.files?.[0] && !preview) return
    setLoading(true); setError('')

    try {
      const formData = new FormData()
      if (fileRef.current?.files?.[0]) {
        formData.append('image', fileRef.current.files[0])
      }
      const r = await api.post('/species/identify', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setResult(r.data)
    } catch (e: any) {
      setError(e.response?.data?.error || '식별에 실패했습니다')
    } finally {
      setLoading(false)
    }
  }

  const confidenceColor = (c: number) =>
    c >= 0.8 ? 'text-green-600' : c >= 0.5 ? 'text-yellow-600' : 'text-gray-500'

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-2">AI 어종 식별</h1>
      <p className="text-gray-600 mb-6">
        물고기 사진을 업로드하면 AI가 어종을 식별합니다.
      </p>

      {/* 이미지 업로드 영역 */}
      <div
        className={`border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-colors
          ${preview ? 'border-blue-300 bg-blue-50' : 'border-gray-300 hover:border-blue-300 hover:bg-blue-50'}`}
        onClick={() => fileRef.current?.click()}
        onDrop={handleDrop}
        onDragOver={e => e.preventDefault()}
      >
        {preview ? (
          <img src={preview} alt="preview" className="max-h-64 mx-auto rounded-lg object-contain" />
        ) : (
          <div>
            <div className="text-5xl mb-3">🐠</div>
            <div className="text-gray-500">클릭하거나 이미지를 드래그하세요</div>
            <div className="text-sm text-gray-400 mt-1">JPG, PNG, WEBP (최대 5MB)</div>
          </div>
        )}
        <input
          ref={fileRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={e => e.target.files?.[0] && handleFile(e.target.files[0])}
        />
      </div>

      {error && <p className="text-red-500 mt-2 text-sm">{error}</p>}

      <button
        onClick={handleIdentify}
        disabled={!preview || loading}
        className="w-full mt-4 py-3 bg-blue-600 text-white rounded-xl font-medium
          hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {loading ? (
          <span className="flex items-center justify-center gap-2">
            <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24" fill="none">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
            </svg>
            AI 분석 중...
          </span>
        ) : 'AI로 어종 식별'}
      </button>

      {/* 결과 */}
      {result && result.candidates.length > 0 && (
        <div className="mt-6">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-lg font-semibold">식별 결과</h2>
            <span className="text-xs text-gray-400">{result.processing_ms}ms</span>
          </div>
          <div className="space-y-3">
            {result.candidates.map((c, i) => (
              <div
                key={i}
                className={`rounded-xl border p-4 ${i === 0 ? 'border-blue-200 bg-blue-50' : 'border-gray-100'}`}
              >
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      {i === 0 && <span className="text-xs bg-blue-600 text-white px-2 py-0.5 rounded-full">최유력</span>}
                      <span className="font-semibold">{c.name}</span>
                    </div>
                    <div className="text-sm text-gray-500 italic">{c.scientific_name}</div>
                  </div>
                  <div className={`text-lg font-bold ${confidenceColor(c.confidence)}`}>
                    {Math.round(c.confidence * 100)}%
                  </div>
                </div>
                <p className="text-sm text-gray-600 mt-2">{c.description}</p>
                {c.fish_data_id && (
                  <Link
                    to={`/fish/${c.fish_data_id}`}
                    className="inline-block mt-2 text-sm text-blue-600 hover:underline"
                  >
                    → 백과사전에서 보기
                  </Link>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {result && result.candidates.length === 0 && (
        <div className="mt-6 text-center text-gray-500">
          <div className="text-3xl mb-2">🤔</div>
          <div>물고기를 식별하지 못했습니다. 더 선명한 사진을 사용해보세요.</div>
        </div>
      )}
    </div>
  )
}
