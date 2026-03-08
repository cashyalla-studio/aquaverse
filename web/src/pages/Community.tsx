import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Link, useParams } from 'react-router-dom'
import { MessageSquare, Eye, Heart, Pin, Lock, Plus } from 'lucide-react'
import { apiClient } from '../api/client'
import { clsx } from 'clsx'

const CATEGORY_ICONS: Record<string, string> = {
  GENERAL: '💬',
  QUESTION: '❓',
  SHOWCASE: '🏆',
  BREEDING: '🐣',
  DISEASES: '💊',
  EQUIPMENT: '⚙️',
  NEWS: '📰',
}

export default function CommunityHome() {
  const { t, i18n } = useTranslation()
  const locale = i18n.language

  const { data: boards, isLoading } = useQuery({
    queryKey: ['boards', locale],
    queryFn: () => apiClient.get('/boards', {
      headers: { 'X-Locale': locale },
    }).then((r) => r.data as Board[]),
    staleTime: 60 * 60 * 1000,
  })

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">{t('communityBoards')}</h1>
        <p className="text-sm text-gray-500 mt-1">
          Board language: <strong>{locale}</strong>
          {' · '}
          <span className="text-xs text-gray-400">
            Content is strictly separated by language
          </span>
        </p>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 7 }).map((_, i) => (
            <div key={i} className="h-16 bg-gray-200 rounded-xl animate-pulse" />
          ))}
        </div>
      ) : (
        <div className="grid gap-3">
          {boards?.map((board) => (
            <Link key={board.id} to={`/community/${board.id}`}>
              <div className="bg-white rounded-xl shadow-sm p-4 hover:shadow-md transition-shadow flex items-center gap-4">
                <span className="text-2xl w-10 text-center">{CATEGORY_ICONS[board.category] ?? '💬'}</span>
                <div className="flex-1">
                  <h3 className="font-semibold text-gray-900">{board.name}</h3>
                  {board.description && (
                    <p className="text-sm text-gray-500 mt-0.5">{board.description}</p>
                  )}
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium text-gray-700">{(board.post_count ?? 0).toLocaleString()}</p>
                  <p className="text-xs text-gray-400">posts</p>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

interface Board {
  id: number
  locale: string
  category: string
  name: string
  description?: string
  is_rtl: boolean
  post_count: number
}

interface Post {
  id: number
  board_id: number
  locale: string
  title: string
  content: string
  view_count: number
  like_count: number
  comment_count: number
  is_pinned: boolean
  is_locked: boolean
  created_at: string
  author_id: string
}

// 게시판 글 목록
export function BoardPage() {
  const { boardID } = useParams<{ boardID: string }>()
  const { t, i18n } = useTranslation()
  const locale = i18n.language

  const { data, isLoading, isError } = useQuery({
    queryKey: ['board-posts', boardID, locale],
    queryFn: () => apiClient.get(`/boards/${boardID}/posts`, {
      headers: { 'X-Locale': locale },
    }).then((r) => r.data as { items: Post[]; total_count: number }),
    enabled: !!boardID,
  })

  if (isError) return (
    <div className="text-center py-12">
      <p className="text-gray-500 text-lg">🚫 This board is not available in your language.</p>
      <p className="text-sm text-gray-400 mt-2">Please switch your language to access this board.</p>
      <Link to="/community" className="text-primary-600 mt-4 inline-block">← Back to boards</Link>
    </div>
  )

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <Link to="/community" className="text-sm text-gray-500 hover:text-primary-600">← {t('communityBoards')}</Link>
          <h1 className="text-xl font-bold text-gray-900 mt-1">Board #{boardID}</h1>
        </div>
        <Link
          to={`/community/${boardID}/new`}
          className="flex items-center gap-2 bg-primary-500 text-white px-4 py-2 rounded-xl text-sm font-medium hover:bg-primary-600"
        >
          <Plus size={16} />
          {t('communityNewPost')}
        </Link>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 10 }).map((_, i) => (
            <div key={i} className="h-14 bg-gray-200 rounded-xl animate-pulse" />
          ))}
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow-sm divide-y">
          {data?.items.map((post) => (
            <Link key={post.id} to={`/community/${boardID}/post/${post.id}`}>
              <div className="p-4 hover:bg-gray-50 transition-colors">
                <div className="flex items-center gap-2">
                  {post.is_pinned && <Pin size={14} className="text-primary-500 flex-shrink-0" />}
                  {post.is_locked && <Lock size={14} className="text-gray-400 flex-shrink-0" />}
                  <h3 className={clsx('text-sm font-medium truncate', post.is_pinned ? 'text-primary-700' : 'text-gray-900')}>
                    {post.title}
                  </h3>
                </div>
                <div className="flex items-center gap-4 mt-1.5 text-xs text-gray-400">
                  <span className="flex items-center gap-1"><Eye size={11} /> {post.view_count}</span>
                  <span className="flex items-center gap-1"><Heart size={11} /> {post.like_count}</span>
                  <span className="flex items-center gap-1"><MessageSquare size={11} /> {post.comment_count}</span>
                  <span>{new Date(post.created_at).toLocaleDateString(locale)}</span>
                </div>
              </div>
            </Link>
          ))}
          {data?.items.length === 0 && (
            <div className="py-12 text-center text-gray-500">{t('common.no_results')}</div>
          )}
        </div>
      )}
    </div>
  )
}
