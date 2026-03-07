import { useEffect, useState } from 'react'
import api from '../../api/client'

interface Business {
  id: number
  store_name: string
  description: string
  city: string
  address: string
  phone: string
  lat: number | null
  lng: number | null
  logo_url: string
  is_verified: boolean
  avg_rating: number
  review_count: number
  distance_km?: number
}

export default function MapPage() {
  const [businesses, setBusinesses] = useState<Business[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [userLat, setUserLat] = useState<number | null>(null)
  const [userLng, setUserLng] = useState<number | null>(null)
  const [radius, setRadius] = useState(5)
  const [geoStatus, setGeoStatus] = useState<'idle' | 'loading' | 'done' | 'denied'>('idle')

  const requestLocation = () => {
    if (!navigator.geolocation) {
      setGeoStatus('denied')
      return
    }
    setGeoStatus('loading')
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setUserLat(pos.coords.latitude)
        setUserLng(pos.coords.longitude)
        setGeoStatus('done')
      },
      () => setGeoStatus('denied'),
      { timeout: 10000 }
    )
  }

  const search = async () => {
    if (!userLat || !userLng) return
    setLoading(true)
    setError('')
    try {
      const res = await api.get('/businesses/nearby', {
        params: { lat: userLat, lng: userLng, radius, limit: 20 }
      })
      setBusinesses(res.data.businesses || [])
    } catch (e: any) {
      setError(e.response?.data?.error || '검색 실패')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (geoStatus === 'done') search()
  }, [geoStatus])

  const openNaverMap = (b: Business) => {
    if (b.lat && b.lng) {
      window.open(`https://map.naver.com/v5/?c=${b.lng},${b.lat},15,0,0,0,dh`, '_blank')
    } else if (b.address) {
      window.open(`https://map.naver.com/v5/search/${encodeURIComponent(b.address)}`, '_blank')
    }
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">주변 수족관 찾기</h1>

      <div className="bg-white rounded-xl shadow-sm border p-4 mb-6">
        <div className="flex items-center gap-4 mb-3">
          <div>
            <label className="text-sm font-medium">검색 반경</label>
            <div className="flex items-center gap-2 mt-1">
              {[3, 5, 10, 20].map(r => (
                <button key={r} onClick={() => setRadius(r)}
                  className={`px-3 py-1 rounded-full text-sm border ${radius === r ? 'bg-blue-600 text-white border-blue-600' : 'text-gray-600'}`}>
                  {r}km
                </button>
              ))}
            </div>
          </div>
        </div>

        {geoStatus === 'idle' && (
          <button onClick={requestLocation}
            className="w-full bg-blue-600 text-white py-2 rounded-lg font-medium">
            내 위치로 검색
          </button>
        )}
        {geoStatus === 'loading' && <p className="text-center text-gray-500">위치 확인 중...</p>}
        {geoStatus === 'done' && (
          <button onClick={search}
            className="w-full bg-green-600 text-white py-2 rounded-lg font-medium">
            다시 검색 ({radius}km 반경)
          </button>
        )}
        {geoStatus === 'denied' && (
          <p className="text-red-500 text-sm text-center">위치 권한이 거부됐습니다. 브라우저 설정을 확인하세요.</p>
        )}
      </div>

      {error && <p className="text-red-500 text-sm mb-4">{error}</p>}

      {loading && <div className="text-center py-8">검색 중...</div>}

      {!loading && businesses.length > 0 && (
        <div className="space-y-3">
          <p className="text-sm text-gray-500">{businesses.length}개 업체 발견</p>
          {businesses.map(b => (
            <div key={b.id} className="bg-white rounded-xl shadow-sm border p-4 flex gap-3">
              <div className="w-14 h-14 rounded-full bg-blue-50 flex items-center justify-center flex-shrink-0 overflow-hidden">
                {b.logo_url ? <img src={b.logo_url} alt={b.store_name} className="w-full h-full object-cover" />
                  : <span className="text-2xl">🐠</span>}
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold">{b.store_name}</h3>
                  {b.is_verified && <span className="text-xs bg-blue-100 text-blue-700 px-1.5 py-0.5 rounded">인증</span>}
                  {b.distance_km != null && (
                    <span className="text-xs text-gray-400 ml-auto">{b.distance_km.toFixed(1)}km</span>
                  )}
                </div>
                {b.address && <p className="text-sm text-gray-500">{b.address}</p>}
                <div className="flex items-center gap-3 mt-1">
                  <span className="text-yellow-400 text-sm">{'★'.repeat(Math.round(b.avg_rating))}</span>
                  {b.phone && <a href={`tel:${b.phone}`} className="text-blue-600 text-sm">전화</a>}
                  <button onClick={() => openNaverMap(b)} className="text-blue-600 text-sm">지도 보기</button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {!loading && geoStatus === 'done' && businesses.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          <p>{radius}km 반경 내 등록된 업체가 없습니다.</p>
          <button onClick={() => setRadius(r => Math.min(r * 2, 50))}
            className="mt-2 text-blue-600 text-sm">반경 넓히기</button>
        </div>
      )}
    </div>
  )
}
