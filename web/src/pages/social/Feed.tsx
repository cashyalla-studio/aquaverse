import { useEffect, useState } from 'react'
import api from '../../api/client'

interface ActivityItem {
  id: number
  actor_id: string
  actor_name: string
  verb: string
  object_type: string
  object_id?: number
  object_data?: Record<string, unknown>
  created_at: string
}

interface Suggestion {
  user_id: string
  username: string
  trust_score: number
  common_fish: string
}

const verbLabel: Record<string, string> = {
  LISTED: '새 분양글을 올렸습니다',
  SOLD: '분양을 완료했습니다',
  REVIEWED: '리뷰를 남겼습니다',
  JOINED: '가입했습니다',
}

export default function SocialFeed() {
  const [feed, setFeed] = useState<ActivityItem[]>([])
  const [suggestions, setSuggestions] = useState<Suggestion[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      api.get('/social/feed'),
      api.get('/social/suggestions'),
    ]).then(([feedRes, suggestRes]) => {
      setFeed(feedRes.data.feed || [])
      setSuggestions(suggestRes.data.suggestions || [])
    }).catch(() => {}).finally(() => setLoading(false))
  }, [])

  const handleFollow = async (userId: string) => {
    try {
      await api.post(`/social/users/${userId}/follow`)
      setSuggestions(prev => prev.filter(s => s.user_id !== userId))
    } catch (err) {
      console.error('follow failed', err)
    }
  }

  return (
    <div className="max-w-2xl mx-auto p-4">
      <h1 className="text-xl font-bold mb-6">팔로잉 피드</h1>

      <div className="flex gap-6">
        {/* 피드 */}
        <div className="flex-1">
          {loading ? (
            <div className="text-gray-500">로딩 중...</div>
          ) : feed.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              <div className="text-4xl mb-2">🐠</div>
              <div>팔로우한 사용자의 활동이 없습니다.</div>
              <div className="text-sm mt-1">오른쪽에서 추천 사용자를 팔로우해보세요!</div>
            </div>
          ) : (
            <div className="space-y-3">
              {feed.map(item => (
                <div key={item.id} className="bg-white border rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center text-sm font-bold text-blue-600">
                      {item.actor_name[0]?.toUpperCase()}
                    </div>
                    <div>
                      <span className="font-medium">{item.actor_name}</span>
                      <span className="text-gray-600 ml-2">{verbLabel[item.verb] || item.verb}</span>
                    </div>
                  </div>
                  {item.object_data && (
                    <div className="text-sm text-gray-500 ml-10">
                      {JSON.stringify(item.object_data)}
                    </div>
                  )}
                  <div className="text-xs text-gray-400 ml-10 mt-1">
                    {new Date(item.created_at).toLocaleDateString('ko-KR', { month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* 팔로우 추천 */}
        {suggestions.length > 0 && (
          <div className="w-64 shrink-0">
            <h2 className="font-medium mb-3 text-gray-700">추천 팔로우</h2>
            <div className="space-y-2">
              {suggestions.map(s => (
                <div key={s.user_id} className="bg-white border rounded-lg p-3 flex items-center justify-between">
                  <div>
                    <div className="font-medium text-sm">{s.username}</div>
                    {s.common_fish && (
                      <div className="text-xs text-gray-500">🐟 {s.common_fish}</div>
                    )}
                    <div className="text-xs text-gray-400">신뢰도 {s.trust_score}</div>
                  </div>
                  <button
                    onClick={() => handleFollow(s.user_id)}
                    className="px-2 py-1 bg-blue-600 text-white text-xs rounded hover:bg-blue-700"
                  >
                    팔로우
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
