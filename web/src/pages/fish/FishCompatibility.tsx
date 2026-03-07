import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import api from '../../api/client'

interface FishRec {
  fish_id: number
  fish_name: string
  scientific_name: string
  image_url: string
  reason: string
  score: number
}

export default function FishCompatibility() {
  const { id } = useParams<{ id: string }>()
  const [compatible, setCompatible] = useState<FishRec[]>([])
  const [fishName, setFishName] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    Promise.all([
      api.get(`/fish/${id}`),
      api.get(`/fish/${id}/compatible`)
    ]).then(([fishRes, compatRes]) => {
      setFishName(fishRes.data.common_name || fishRes.data.name || '')
      setCompatible(compatRes.data.compatible_fish || [])
    }).finally(() => setLoading(false))
  }, [id])

  if (loading) return <div className="p-8 text-center">로딩 중...</div>

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="mb-6">
        <Link to={`/fish/${id}`} className="text-blue-500 hover:underline text-sm">
          &larr; {fishName} 상세로 돌아가기
        </Link>
        <h1 className="text-2xl font-bold mt-2">{fishName} &mdash; 합사 가능 어종</h1>
        <p className="text-gray-500 text-sm mt-1">총 {compatible.length}종</p>
      </div>

      {compatible.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          <p>등록된 호환성 데이터가 없습니다.</p>
          <p className="text-sm mt-1">Rule-based 추천은 수질 파라미터 데이터가 필요합니다.</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {compatible.map((fish) => (
            <Link
              key={fish.fish_id}
              to={`/fish/${fish.fish_id}`}
              className="bg-white rounded-xl shadow-sm border hover:shadow-md transition-shadow p-4 flex flex-col items-center text-center"
            >
              {fish.image_url ? (
                <img
                  src={fish.image_url}
                  alt={fish.fish_name}
                  className="w-20 h-20 object-cover rounded-full mb-3"
                />
              ) : (
                <div className="w-20 h-20 bg-blue-50 rounded-full mb-3 flex items-center justify-center text-3xl">
                  🐟
                </div>
              )}
              <p className="font-semibold text-sm">{fish.fish_name}</p>
              {fish.scientific_name && (
                <p className="text-xs text-gray-400 italic mt-0.5">{fish.scientific_name}</p>
              )}
              {fish.reason && (
                <p className="text-xs text-green-600 mt-2 bg-green-50 rounded px-2 py-1">
                  {fish.reason}
                </p>
              )}
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
