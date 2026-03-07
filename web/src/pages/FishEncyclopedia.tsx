import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { Search, Filter, Droplets, Thermometer, Ruler } from 'lucide-react'
import { fishApi, type FishListItem } from '../api/fish'
import { clsx } from 'clsx'

const CARE_LEVELS = ['BEGINNER', 'INTERMEDIATE', 'EXPERT'] as const
const CARE_COLORS = {
  BEGINNER: 'bg-green-100 text-green-700',
  INTERMEDIATE: 'bg-yellow-100 text-yellow-700',
  EXPERT: 'bg-red-100 text-red-700',
}

export default function FishEncyclopedia() {
  const { t, i18n } = useTranslation()
  const [search, setSearch] = useState('')
  const [careLevel, setCareLevel] = useState('')
  const [family, setFamily] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading, isError } = useQuery({
    queryKey: ['fish', { search, careLevel, family, page, locale: i18n.language }],
    queryFn: () =>
      fishApi.list({ q: search, care_level: careLevel, family, page, limit: 24, locale: i18n.language })
        .then((r) => r.data),
    staleTime: 5 * 60 * 1000,
  })

  const { data: families } = useQuery({
    queryKey: ['fish-families'],
    queryFn: () => fishApi.families().then((r) => r.data),
    staleTime: 60 * 60 * 1000,
  })

  return (
    <div>
      {/* 헤더 */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">{t('nav.encyclopedia')}</h1>
        <p className="text-gray-500 mt-1">
          {data?.total_count.toLocaleString()} species
        </p>
      </div>

      {/* 검색 + 필터 */}
      <div className="bg-white rounded-xl shadow-sm p-4 mb-6 flex flex-wrap gap-3">
        {/* 검색 */}
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
          <input
            type="text"
            placeholder={t('common.search')}
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-300"
          />
        </div>

        {/* 난이도 필터 */}
        <select
          value={careLevel}
          onChange={(e) => { setCareLevel(e.target.value); setPage(1) }}
          className="border border-gray-200 rounded-lg px-3 py-2 text-sm bg-white"
        >
          <option value="">{t('fish.care_level')}: All</option>
          {CARE_LEVELS.map((l) => (
            <option key={l} value={l}>{t(`fish.care_levels.${l}`)}</option>
          ))}
        </select>

        {/* 과(Family) 필터 */}
        <select
          value={family}
          onChange={(e) => { setFamily(e.target.value); setPage(1) }}
          className="border border-gray-200 rounded-lg px-3 py-2 text-sm bg-white"
        >
          <option value="">Family: All</option>
          {families?.map((f) => <option key={f} value={f}>{f}</option>)}
        </select>
      </div>

      {/* 그리드 */}
      {isLoading && (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
          {Array.from({ length: 12 }).map((_, i) => (
            <div key={i} className="bg-white rounded-xl shadow-sm overflow-hidden animate-pulse">
              <div className="h-40 bg-gray-200" />
              <div className="p-3 space-y-2">
                <div className="h-4 bg-gray-200 rounded w-3/4" />
                <div className="h-3 bg-gray-200 rounded w-1/2" />
              </div>
            </div>
          ))}
        </div>
      )}

      {isError && (
        <div className="text-center py-12 text-gray-500">{t('common.error')}</div>
      )}

      {data && (
        <>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
            {data.items.map((fish) => (
              <FishCard key={fish.id} fish={fish} />
            ))}
          </div>

          {data.items.length === 0 && (
            <div className="text-center py-12 text-gray-500">{t('common.no_results')}</div>
          )}

          {/* 페이지네이션 */}
          {data.total_count > 24 && (
            <div className="mt-8 flex justify-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-4 py-2 rounded-lg border text-sm disabled:opacity-50 hover:bg-gray-50"
              >
                ←
              </button>
              <span className="px-4 py-2 text-sm text-gray-600">
                {t('common.page')} {page} {t('common.of')} {Math.ceil(data.total_count / 24)}
              </span>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page >= Math.ceil(data.total_count / 24)}
                className="px-4 py-2 rounded-lg border text-sm disabled:opacity-50 hover:bg-gray-50"
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

function FishCard({ fish }: { fish: FishListItem }) {
  const { t } = useTranslation()

  return (
    <Link to={`/fish/${fish.id}`} className="group">
      <div className="bg-white rounded-xl shadow-sm overflow-hidden hover:shadow-md transition-shadow">
        {/* 이미지 */}
        <div className="relative h-40 bg-gradient-to-br from-blue-50 to-cyan-100 overflow-hidden">
          {fish.primary_image_url ? (
            <img
              src={fish.primary_image_url}
              alt={fish.common_name}
              className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
              loading="lazy"
            />
          ) : (
            <div className="flex items-center justify-center h-full text-4xl">🐠</div>
          )}
          {fish.care_level && (
            <span className={clsx(
              'absolute top-2 right-2 text-xs px-2 py-0.5 rounded-full font-medium',
              CARE_COLORS[fish.care_level as keyof typeof CARE_COLORS],
            )}>
              {t(`fish.care_levels.${fish.care_level}`)}
            </span>
          )}
        </div>

        {/* 정보 */}
        <div className="p-3">
          <h3 className="font-semibold text-gray-900 text-sm truncate">
            {fish.common_name}
          </h3>
          <p className="text-xs text-gray-400 italic truncate mt-0.5">
            {fish.scientific_name}
          </p>

          {/* 수치 정보 */}
          <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1">
            {fish.max_size_cm && (
              <span className="flex items-center gap-1 text-xs text-gray-500">
                <Ruler size={11} /> {fish.max_size_cm}cm
              </span>
            )}
            {fish.min_tank_size_liters && (
              <span className="flex items-center gap-1 text-xs text-gray-500">
                <Droplets size={11} /> {fish.min_tank_size_liters}L
              </span>
            )}
          </div>
        </div>
      </div>
    </Link>
  )
}
