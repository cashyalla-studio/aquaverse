import { useEffect, useRef, useState, useCallback } from 'react'
import api from '../../api/client'
import { useAuthStore } from '../../store/authStore'

interface VideoPost {
  id: number; title: string; description: string
  video_url: string; thumbnail_url: string
  username: string; view_count: number; like_count: number
  is_liked: boolean; duration_sec: number; created_at: string
}

function VideoCard({ video, onLike }: { video: VideoPost; onLike: (id: number) => void }) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const cardRef = useRef<HTMLDivElement>(null)
  const [liked, setLiked] = useState(video.is_liked)
  const [likeCount, setLikeCount] = useState(video.like_count)

  // Intersection Observer로 자동 재생/일시정지
  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          videoRef.current?.play().catch(() => {})
          api.post(`/videos/${video.id}/view`).catch(() => {})
        } else {
          videoRef.current?.pause()
        }
      },
      { threshold: 0.7 }
    )
    if (cardRef.current) observer.observe(cardRef.current)
    return () => observer.disconnect()
  }, [video.id])

  const handleLike = async () => {
    const newLiked = !liked
    setLiked(newLiked)
    setLikeCount(c => newLiked ? c + 1 : c - 1)
    try {
      await api.post(`/videos/${video.id}/like`)
    } catch {
      setLiked(!newLiked)
      setLikeCount(c => newLiked ? c - 1 : c + 1)
    }
    onLike(video.id)
  }

  return (
    <div ref={cardRef} className="relative bg-black rounded-xl overflow-hidden" style={{ height: '70vh', maxHeight: 600 }}>
      {video.video_url ? (
        <video
          ref={videoRef}
          src={video.video_url}
          poster={video.thumbnail_url}
          className="w-full h-full object-cover"
          loop muted playsInline
          onClick={() => videoRef.current?.paused ? videoRef.current?.play() : videoRef.current?.pause()}
        />
      ) : (
        <div className="w-full h-full bg-gray-800 flex items-center justify-center">
          <span className="text-white text-4xl">🐠</span>
        </div>
      )}

      {/* 오버레이 */}
      <div className="absolute bottom-0 left-0 right-0 p-4 bg-gradient-to-t from-black/80 to-transparent text-white">
        <p className="font-semibold text-sm">@{video.username}</p>
        <p className="text-sm mt-0.5">{video.title}</p>
        {video.description && <p className="text-xs text-gray-300 mt-0.5 line-clamp-2">{video.description}</p>}
      </div>

      {/* 우측 액션 버튼 */}
      <div className="absolute right-3 bottom-16 flex flex-col items-center gap-4">
        <button onClick={handleLike} className="flex flex-col items-center text-white">
          <span className={`text-2xl ${liked ? 'text-red-400' : ''}`}>{liked ? '❤️' : '🤍'}</span>
          <span className="text-xs mt-0.5">{likeCount}</span>
        </button>
        <div className="flex flex-col items-center text-white">
          <span className="text-xl">👁</span>
          <span className="text-xs mt-0.5">{video.view_count}</span>
        </div>
      </div>
    </div>
  )
}

export default function VideoFeed() {
  const { user } = useAuthStore()
  const [videos, setVideos] = useState<VideoPost[]>([])
  const [loading, setLoading] = useState(true)
  const [offset, setOffset] = useState(0)
  const [hasMore, setHasMore] = useState(true)

  const loadMore = useCallback(async () => {
    if (!hasMore) return
    try {
      const res = await api.get('/videos', { params: { limit: 5, offset } })
      const newVids = res.data.videos || []
      setVideos(v => [...v, ...newVids])
      setOffset(o => o + newVids.length)
      setHasMore(newVids.length === 5)
    } finally {
      setLoading(false)
    }
  }, [offset, hasMore])

  useEffect(() => { loadMore() }, [])

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, scrollHeight, clientHeight } = e.currentTarget
    if (scrollHeight - scrollTop - clientHeight < 300 && !loading) {
      loadMore()
    }
  }, [loading, loadMore])

  return (
    <div className="max-w-md mx-auto">
      <div className="flex items-center justify-between p-4">
        <h1 className="text-xl font-bold">수조 영상</h1>
        {user && (
          <button className="text-sm text-blue-600 font-medium">+ 업로드</button>
        )}
      </div>

      {loading && videos.length === 0 ? (
        <div className="text-center py-20">로딩 중...</div>
      ) : videos.length === 0 ? (
        <div className="text-center py-20 text-gray-500">
          <p className="text-4xl mb-4">🎬</p>
          <p>아직 영상이 없습니다.</p>
          <p className="text-sm mt-1">첫 수조 영상을 공유해보세요!</p>
        </div>
      ) : (
        <div
          className="space-y-4 px-4 pb-8 overflow-y-auto"
          style={{ maxHeight: 'calc(100vh - 120px)' }}
          onScroll={handleScroll}
        >
          {videos.map(v => (
            <VideoCard key={v.id} video={v} onLike={() => {}} />
          ))}
          {loading && <div className="text-center py-4 text-gray-500">로딩 중...</div>}
          {!hasMore && <div className="text-center py-4 text-gray-400 text-sm">모두 로드됨</div>}
        </div>
      )}
    </div>
  )
}
