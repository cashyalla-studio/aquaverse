import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import api from '../../api/client'
import { useAuthStore } from '../../store/authStore'

interface Business {
  id: number; store_name: string; description: string
  address: string; city: string; phone: string; website: string
  logo_url: string; is_verified: boolean; avg_rating: number; review_count: number
}
interface Review {
  id: number; reviewer_name: string; rating: number; content: string; created_at: string
}

export default function BusinessDetail() {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuthStore()
  const [business, setBusiness] = useState<Business | null>(null)
  const [reviews, setReviews] = useState<Review[]>([])
  const [reviewForm, setReviewForm] = useState({ rating: 5, content: '' })
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!id) return
    Promise.all([
      api.get(`/businesses/${id}`),
      api.get(`/businesses/${id}/reviews`),
    ]).then(([b, r]) => {
      setBusiness(b.data)
      setReviews(r.data.reviews || [])
    })
  }, [id])

  const submitReview = async () => {
    setSubmitting(true)
    try {
      const r = await api.post(`/businesses/${id}/reviews`, reviewForm)
      setReviews(prev => [r.data, ...prev])
      setReviewForm({ rating: 5, content: '' })
    } finally {
      setSubmitting(false)
    }
  }

  if (!business) return <div className="p-8 text-center">로딩 중...</div>

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="bg-white rounded-xl shadow-sm border p-6 mb-6">
        <div className="flex items-start gap-4">
          <div className="w-20 h-20 rounded-full bg-blue-50 flex items-center justify-center overflow-hidden flex-shrink-0">
            {business.logo_url
              ? <img src={business.logo_url} alt={business.store_name} className="w-full h-full object-cover" />
              : <span className="text-3xl">🐠</span>}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold">{business.store_name}</h1>
              {business.is_verified && <span className="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded-full">인증업체</span>}
            </div>
            <p className="text-gray-500 text-sm mt-1">{business.city} {business.address}</p>
            <div className="flex items-center gap-1 mt-1">
              <span className="text-yellow-500">{'★'.repeat(Math.round(business.avg_rating))}</span>
              <span className="text-sm text-gray-500">{business.avg_rating.toFixed(1)} ({business.review_count}개 리뷰)</span>
            </div>
          </div>
        </div>
        {business.description && <p className="mt-4 text-gray-700">{business.description}</p>}
        <div className="mt-4 flex flex-wrap gap-3">
          {business.phone && <a href={`tel:${business.phone}`} className="text-blue-600 text-sm">📞 {business.phone}</a>}
          {business.website && <a href={business.website} target="_blank" rel="noopener noreferrer" className="text-blue-600 text-sm">🌐 웹사이트</a>}
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border p-6">
        <h2 className="font-semibold mb-4">리뷰 ({reviews.length})</h2>

        {user && (
          <div className="mb-6 p-4 bg-gray-50 rounded-lg">
            <div className="flex gap-1 mb-2">
              {[1,2,3,4,5].map(n => (
                <button key={n} onClick={() => setReviewForm(f => ({...f, rating: n}))}
                  className={`text-2xl ${n <= reviewForm.rating ? 'text-yellow-400' : 'text-gray-300'}`}>★</button>
              ))}
            </div>
            <textarea
              className="w-full border rounded-lg p-2 text-sm resize-none"
              rows={3} placeholder="리뷰를 작성해주세요..."
              value={reviewForm.content}
              onChange={e => setReviewForm(f => ({...f, content: e.target.value}))}
            />
            <button onClick={submitReview} disabled={submitting}
              className="mt-2 bg-blue-600 text-white px-4 py-1.5 rounded-lg text-sm disabled:opacity-50">
              {submitting ? '제출 중...' : '리뷰 등록'}
            </button>
          </div>
        )}

        <div className="space-y-4">
          {reviews.map(r => (
            <div key={r.id} className="border-b pb-4 last:border-0">
              <div className="flex items-center justify-between">
                <span className="font-medium text-sm">{r.reviewer_name}</span>
                <span className="text-yellow-400">{'★'.repeat(r.rating)}</span>
              </div>
              {r.content && <p className="text-sm text-gray-700 mt-1">{r.content}</p>}
              <p className="text-xs text-gray-400 mt-1">{new Date(r.created_at).toLocaleDateString('ko-KR')}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
