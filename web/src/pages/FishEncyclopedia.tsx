import { useState, useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { Search, Sliders, Droplets, Ruler, Sparkles } from 'lucide-react'
import { fishApi, type FishListItem, type CreatureCategory } from '../api/fish'
import { clsx } from 'clsx'
import { events } from '../lib/posthog'
import CareToday from '../components/CareToday'

const CARE_LEVELS = ['BEGINNER', 'INTERMEDIATE', 'EXPERT'] as const
const CARE_COLORS = {
  BEGINNER:     'bg-emerald-100 text-emerald-700',
  INTERMEDIATE: 'bg-amber-100 text-amber-700',
  EXPERT:       'bg-red-100 text-red-700',
}
const CARE_LABELS = {
  BEGINNER:     '입문',
  INTERMEDIATE: '중급',
  EXPERT:       '고급',
}

export default function FishEncyclopedia() {
  const { t, i18n } = useTranslation()
  const [search, setSearch] = useState('')
  const [careLevel, setCareLevel] = useState('')
  const [page, setPage] = useState(1)
  const [activeCategory, setActiveCategory] = useState('')
  const [showFilter, setShowFilter] = useState(false)
  const prevSearchRef = useRef('')

  const { data, isLoading, isError } = useQuery({
    queryKey: ['fish', { search, careLevel, page, locale: i18n.language, activeCategory }],
    queryFn: () =>
      fishApi
        .list({ q: search, care_level: careLevel, page, limit: 24, locale: i18n.language, category: activeCategory || undefined })
        .then((r) => r.data),
    staleTime: 5 * 60 * 1000,
  })

  const { data: categories } = useQuery({
    queryKey: ['fish-categories'],
    queryFn: () => fishApi.categories().then((r) => r.data),
    staleTime: 60 * 60 * 1000,
  })

  useEffect(() => {
    if (data && search && search !== prevSearchRef.current) {
      prevSearchRef.current = search
      events.searchPerformed(search, data.items.length)
    }
  }, [data, search])

  useEffect(() => { setPage(1) }, [activeCategory])

  const totalPages = data ? Math.ceil(data.total_count / 24) : 1

  return (
    <div className="space-y-6">
      {/* ── 오늘 케어 배너 (로그인 + 남은 일정 있을 때만 표시) ── */}
      <CareToday />

      {/* ── 히어로 헤더 ── */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary-600 via-cyan-600 to-teal-600 p-6 sm:p-8 text-white">
        <div className="absolute inset-0 opacity-10"
          style={{ backgroundImage: 'radial-gradient(circle at 20% 50%, white 1px, transparent 1px), radial-gradient(circle at 80% 20%, white 1px, transparent 1px)', backgroundSize: '40px 40px' }}
        />
        <div className="relative">
          <div className="flex items-center gap-2 mb-1">
            <Sparkles size={16} className="text-cyan-200" />
            <span className="text-cyan-200 text-sm font-medium">생물 백과사전</span>
          </div>
          <h1 className="text-3xl font-black tracking-tight mb-2">모든 생물을 탐색하세요</h1>
          <p className="text-cyan-100 text-sm mb-5">
            {data?.total_count
              ? `${data.total_count.toLocaleString()}종의 생물 정보`
              : '어종부터 파충류, 곤충까지'}
          </p>

          {/* 검색바 */}
          <div className="relative">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
            <input
              type="text"
              placeholder="어종명, 학명으로 검색..."
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1) }}
              className="w-full pl-11 pr-16 py-3 rounded-xl bg-white text-gray-900 text-sm
                         placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-white/50 shadow-lg"
            />
            <button
              onClick={() => setShowFilter(!showFilter)}
              className={clsx(
                'absolute right-2 top-1/2 -translate-y-1/2 p-2 rounded-lg transition-colors',
                showFilter ? 'bg-primary-100 text-primary-600' : 'text-gray-400 hover:text-gray-600',
              )}
              aria-label="필터"
            >
              <Sliders size={16} />
            </button>
          </div>
        </div>
      </div>

      {/* ── 필터 패널 ── */}
      {showFilter && (
        <div className="card p-4 flex flex-wrap gap-3 items-center">
          <span className="text-sm font-semibold text-gray-700">필터</span>
          <select
            value={careLevel}
            onChange={(e) => { setCareLevel(e.target.value); setPage(1) }}
            className="border border-gray-200 rounded-xl px-3 py-2 text-sm bg-white focus:outline-none focus:ring-2 focus:ring-primary-300"
          >
            <option value="">난이도 전체</option>
            {CARE_LEVELS.map((l) => (
              <option key={l} value={l}>{CARE_LABELS[l]}</option>
            ))}
          </select>
          {(search || careLevel) && (
            <button
              onClick={() => { setSearch(''); setCareLevel(''); setPage(1) }}
              className="text-sm text-red-500 hover:underline"
            >
              초기화
            </button>
          )}
        </div>
      )}

      {/* ── 카테고리 탭 ── */}
      <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-hide -mx-1 px-1">
        {[{ code: '', icon_emoji: '🌍', name_ko: '전체' }, ...(categories ?? [])].map((cat: { code: string; icon_emoji: string; name_ko: string }) => (
          <button
            key={cat.code}
            onClick={() => setActiveCategory(cat.code)}
            className={clsx(
              'flex-shrink-0 flex items-center gap-1.5 px-4 py-2 rounded-full text-sm font-medium transition-all',
              activeCategory === cat.code
                ? 'bg-primary-600 text-white shadow-sm shadow-primary-200'
                : 'bg-white text-gray-600 border border-gray-200 hover:border-primary-300 hover:text-primary-600',
            )}
          >
            <span className="text-base leading-none">{cat.icon_emoji}</span>
            {cat.name_ko}
          </button>
        ))}
      </div>

      {/* ── 결과 그리드 ── */}
      {isLoading && (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
          {Array.from({ length: 12 }).map((_, i) => (
            <div key={i} className="card overflow-hidden animate-pulse">
              <div className="h-44 bg-gray-100" />
              <div className="p-3 space-y-2">
                <div className="h-4 bg-gray-100 rounded-lg w-3/4" />
                <div className="h-3 bg-gray-100 rounded-lg w-1/2" />
              </div>
            </div>
          ))}
        </div>
      )}

      {isError && (
        <div className="text-center py-16">
          <p className="text-4xl mb-3">🌊</p>
          <p className="text-gray-500">{t('common.error')}</p>
        </div>
      )}

      {data && (
        <>
          {data.items.length > 0 ? (
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
              {data.items.map((fish, idx) => (
                <FishCard key={fish.id} fish={fish} priority={idx < 4} />
              ))}
            </div>
          ) : (
            <div className="text-center py-16">
              <p className="text-5xl mb-3">🔍</p>
              <p className="text-gray-500 font-medium">검색 결과가 없습니다</p>
              <p className="text-gray-400 text-sm mt-1">다른 검색어를 입력해 보세요</p>
            </div>
          )}

          {/* 페이지네이션 */}
          {totalPages > 1 && (
            <div className="flex justify-center items-center gap-2 pt-4">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="btn-secondary px-3 py-2 disabled:opacity-40"
              >
                ←
              </button>
              <div className="flex items-center gap-1">
                {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                  const p = page <= 3 ? i + 1 : page + i - 2
                  if (p < 1 || p > totalPages) return null
                  return (
                    <button
                      key={p}
                      onClick={() => setPage(p)}
                      className={clsx(
                        'w-9 h-9 rounded-xl text-sm font-medium transition-colors',
                        p === page
                          ? 'bg-primary-600 text-white'
                          : 'text-gray-600 hover:bg-gray-100',
                      )}
                    >
                      {p}
                    </button>
                  )
                })}
              </div>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
                className="btn-secondary px-3 py-2 disabled:opacity-40"
              >
                →
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}

function FishCard({ fish, priority }: { fish: FishListItem; priority?: boolean }) {
  return (
    <Link to={`/fish/${fish.id}`} className="group">
      <div className="card-hover overflow-hidden">
        {/* 이미지 */}
        <div className="relative h-44 bg-gradient-to-br from-blue-50 via-cyan-50 to-teal-50 overflow-hidden">
          {fish.primary_image_url ? (
            <img
              src={fish.primary_image_url}
              alt={fish.common_name}
              className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
              loading={priority ? 'eager' : 'lazy'}
            />
          ) : (
            <div className="flex items-center justify-center h-full text-5xl opacity-60">
              {getCategoryEmoji(fish)}
            </div>
          )}

          {/* 난이도 배지 */}
          {fish.care_level && (
            <span className={clsx(
              'absolute top-2 right-2 text-xs px-2 py-0.5 rounded-full font-semibold',
              CARE_COLORS[fish.care_level as keyof typeof CARE_COLORS],
            )}>
              {CARE_LABELS[fish.care_level as keyof typeof CARE_LABELS] ?? fish.care_level}
            </span>
          )}

          {/* 호버 오버레이 */}
          <div className="absolute inset-0 bg-gradient-to-t from-black/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-end p-3">
            <span className="text-white text-xs font-medium">상세 보기 →</span>
          </div>
        </div>

        {/* 정보 */}
        <div className="p-3">
          <h3 className="font-semibold text-gray-900 text-sm truncate leading-tight">
            {fish.common_name}
          </h3>
          <p className="text-xs text-gray-400 italic truncate mt-0.5">
            {fish.scientific_name}
          </p>

          {/* 수치 */}
          <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1">
            {fish.max_size_cm && (
              <span className="flex items-center gap-0.5 text-xs text-gray-400">
                <Ruler size={10} className="flex-shrink-0" />
                {fish.max_size_cm}cm
              </span>
            )}
            {fish.min_tank_size_liters && (
              <span className="flex items-center gap-0.5 text-xs text-gray-400">
                <Droplets size={10} className="flex-shrink-0" />
                {fish.min_tank_size_liters}L
              </span>
            )}
          </div>
        </div>
      </div>
    </Link>
  )
}

function getCategoryEmoji(fish: FishListItem): string {
  const cat = (fish as any).creature_category ?? 'fish'
  const map: Record<string, string> = {
    fish: '🐠',
    reptile: '🦎',
    amphibian: '🐸',
    insect: '🐜',
    arachnid: '🕷️',
    bird: '🦜',
    mammal: '🐹',
  }
  return map[cat] ?? '🐠'
}
