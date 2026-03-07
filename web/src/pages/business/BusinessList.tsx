import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import api from '../../api/client'

interface Business {
  id: number
  store_name: string
  description: string
  city: string
  phone: string
  logo_url: string
  is_verified: boolean
  avg_rating: number
  review_count: number
}

export default function BusinessList() {
  const [businesses, setBusinesses] = useState<Business[]>([])
  const [city, setCity] = useState('')
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    api.get('/businesses', { params: { city, limit: 20 } })
      .then(r => setBusinesses(r.data.businesses || []))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">주변 수족관 업체</h1>
      <div className="flex gap-2 mb-6">
        <input
          className="border rounded-lg px-3 py-2 flex-1"
          placeholder="도시 검색 (예: 서울)"
          value={city}
          onChange={e => setCity(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && load()}
        />
        <button onClick={load} className="bg-blue-600 text-white px-4 py-2 rounded-lg">검색</button>
      </div>

      {loading ? (
        <div className="text-center py-12">로딩 중...</div>
      ) : businesses.length === 0 ? (
        <div className="text-center py-12 text-gray-500">등록된 업체가 없습니다.</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {businesses.map(b => (
            <Link key={b.id} to={`/businesses/${b.id}`}
              className="bg-white rounded-xl shadow-sm border hover:shadow-md transition p-4 flex gap-4">
              <div className="w-16 h-16 rounded-full bg-blue-50 flex items-center justify-center flex-shrink-0 overflow-hidden">
                {b.logo_url
                  ? <img src={b.logo_url} alt={b.store_name} className="w-full h-full object-cover" />
                  : <span className="text-2xl">🐠</span>}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold truncate">{b.store_name}</h3>
                  {b.is_verified && <span className="text-xs bg-blue-100 text-blue-700 px-1.5 py-0.5 rounded">인증</span>}
                </div>
                {b.city && <p className="text-sm text-gray-500">{b.city}</p>}
                {b.description && <p className="text-sm text-gray-600 truncate mt-1">{b.description}</p>}
                <div className="flex items-center gap-2 mt-1">
                  <span className="text-yellow-500 text-sm">{'★'.repeat(Math.round(b.avg_rating))}{'☆'.repeat(5 - Math.round(b.avg_rating))}</span>
                  <span className="text-xs text-gray-400">({b.review_count}개)</span>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
