import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { MapPin, Shield, Star, Package, AlertTriangle, Bell, Plus } from 'lucide-react'
import { apiClient } from '../api/client'
import { clsx } from 'clsx'
import type { Listing } from '../types/marketplace'

const HEALTH_COLORS = {
  EXCELLENT: 'text-green-600 bg-green-50',
  GOOD: 'text-blue-600 bg-blue-50',
  DISEASE_HISTORY: 'text-yellow-600 bg-yellow-50',
  UNDER_TREATMENT: 'text-red-600 bg-red-50',
}

const TRADE_ICONS = {
  DIRECT: '🤝',
  COURIER: '📦',
  AQUA_COURIER: '🐟📦',
  ALL: '✅',
}

export default function Marketplace() {
  const { t, i18n } = useTranslation()
  const [search, setSearch] = useState('')
  const [tradeType, setTradeType] = useState('')
  const [page, setPage] = useState(1)
  const [useLocation, setUseLocation] = useState(false)
  const [userLat, setUserLat] = useState<number | null>(null)
  const [userLng, setUserLng] = useState<number | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['listings', { search, tradeType, page, userLat, userLng }],
    queryFn: () => apiClient.get('/listings', {
      params: {
        q: search || undefined,
        trade_type: tradeType || undefined,
        page,
        limit: 20,
        lat: userLat ?? undefined,
        lng: userLng ?? undefined,
        radius_km: userLat ? 50 : undefined,
      },
    }).then((r) => r.data as { items: Listing[]; total_count: number; page: number }),
  })

  const handleGetLocation = () => {
    navigator.geolocation.getCurrentPosition((pos) => {
      setUserLat(pos.coords.latitude)
      setUserLng(pos.coords.longitude)
      setUseLocation(true)
    })
  }

  return (
    <div>
      {/* 헤더 */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{t('marketplaceTitle')}</h1>
          <p className="text-gray-500 text-sm mt-0.5">
            {data?.total_count.toLocaleString()} listings
          </p>
        </div>
        <Link
          to="/marketplace/create"
          className="flex items-center gap-2 bg-primary-500 text-white px-4 py-2 rounded-xl text-sm font-medium hover:bg-primary-600 shadow-sm"
        >
          <Plus size={16} />
          {t('marketplaceCreateListing')}
        </Link>
      </div>

      {/* 필터 바 */}
      <div className="bg-white rounded-xl shadow-sm p-4 mb-6 flex flex-wrap gap-3 items-center">
        <input
          type="text"
          placeholder={t('common.search')}
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          className="flex-1 min-w-[180px] border border-gray-200 rounded-lg px-3 py-2 text-sm"
        />

        <select
          value={tradeType}
          onChange={(e) => { setTradeType(e.target.value); setPage(1) }}
          className="border border-gray-200 rounded-lg px-3 py-2 text-sm bg-white"
        >
          <option value="">All trade types</option>
          <option value="DIRECT">{t('marketplaceDirect')}</option>
          <option value="COURIER">{t('marketplaceCourier')}</option>
          <option value="AQUA_COURIER">{t('marketplaceAquaCourier')}</option>
        </select>

        {/* 위치 기반 토글 */}
        <button
          onClick={useLocation ? () => { setUserLat(null); setUserLng(null); setUseLocation(false) } : handleGetLocation}
          className={clsx(
            'flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm border transition-colors',
            useLocation
              ? 'bg-primary-50 border-primary-200 text-primary-600'
              : 'border-gray-200 text-gray-600 hover:bg-gray-50',
          )}
        >
          <MapPin size={15} />
          {useLocation ? 'Near me' : 'All regions'}
        </button>

        {/* 알림 구독 */}
        <button className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm border border-gray-200 text-gray-600 hover:bg-gray-50">
          <Bell size={15} />
          {t('marketplaceWatchAlert')}
        </button>
      </div>

      {/* 계절 배송 경고 */}
      <div className="mb-4 p-3 bg-amber-50 border border-amber-200 rounded-xl flex items-center gap-2 text-sm text-amber-700">
        <AlertTriangle size={16} />
        Summer shipping risk: High temperatures may harm live fish during transit. Consider using AquaCourier.
      </div>

      {/* 목록 */}
      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="bg-white rounded-xl shadow-sm p-4 animate-pulse flex gap-4">
              <div className="w-24 h-24 bg-gray-200 rounded-xl flex-shrink-0" />
              <div className="flex-1 space-y-2">
                <div className="h-5 bg-gray-200 rounded w-1/2" />
                <div className="h-4 bg-gray-200 rounded w-1/3" />
                <div className="h-6 bg-gray-200 rounded w-24" />
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="space-y-3">
          {data?.items.map((listing) => (
            <ListingCard key={listing.id} listing={listing} />
          ))}
          {data?.items.length === 0 && (
            <div className="text-center py-12 text-gray-500">{t('common.no_results')}</div>
          )}
        </div>
      )}
    </div>
  )
}

function ListingCard({ listing }: { listing: Listing }) {
  const { t } = useTranslation()

  const priceDisplay = listing.price === '0' || listing.price === 0
    ? <span className="text-green-600 font-bold">{t('marketplaceFree')}</span>
    : <span className="text-gray-900 font-bold">{Number(listing.price).toLocaleString()} {listing.currency}</span>

  return (
    <Link to={`/marketplace/${listing.id}`}>
      <div className="bg-white rounded-xl shadow-sm p-4 hover:shadow-md transition-shadow flex gap-4">
        {/* 이미지 */}
        <div className="w-24 h-24 rounded-xl bg-gray-100 flex-shrink-0 overflow-hidden">
          {listing.image_urls?.[0] ? (
            <img src={listing.image_urls[0]} alt={listing.common_name} className="w-full h-full object-cover" />
          ) : (
            <div className="flex items-center justify-center h-full text-3xl">🐠</div>
          )}
        </div>

        {/* 정보 */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2">
            <h3 className="font-semibold text-gray-900 truncate">{listing.title}</h3>
            {priceDisplay}
          </div>

          <p className="text-sm text-gray-500 italic truncate mt-0.5">{listing.common_name}</p>

          {/* 배지 행 */}
          <div className="flex flex-wrap gap-1.5 mt-2">
            {/* 건강 상태 */}
            <span className={clsx('text-xs px-2 py-0.5 rounded-full', HEALTH_COLORS[listing.health_status as keyof typeof HEALTH_COLORS])}>
              {t(`marketplaceHealth${listing.health_status.split('_').map((w: string) => w[0] + w.slice(1).toLowerCase()).join('')}`)}
            </span>

            {/* 거래 방식 */}
            <span className="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-600">
              {TRADE_ICONS[listing.trade_type as keyof typeof TRADE_ICONS]} {t(`marketplace${listing.trade_type === 'AQUA_COURIER' ? 'AquaCourier' : listing.trade_type.charAt(0) + listing.trade_type.slice(1).toLowerCase()}`)}
            </span>

            {/* 직접 번식 */}
            {listing.bred_by_seller && (
              <span className="text-xs px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700">
                🌿 {t('marketplaceBredBySeller')}
              </span>
            )}

            {/* 국제 분양 */}
            {listing.allow_international && (
              <span className="text-xs px-2 py-0.5 rounded-full bg-blue-50 text-blue-600">
                🌍 International
              </span>
            )}
          </div>

          {/* 하단: 위치 + 신뢰도 + 거리 */}
          <div className="flex items-center gap-3 mt-2 text-xs text-gray-400">
            <span className="flex items-center gap-0.5">
              <MapPin size={11} /> {listing.location_text}
            </span>
            {listing.distance_km !== null && listing.distance_km !== undefined && (
              <span>{listing.distance_km.toFixed(1)} km</span>
            )}
            <span className="flex items-center gap-0.5">
              <Shield size={11} />
              {listing.seller_trust_score?.toFixed(1) ?? '—'}
              <Star size={9} className="text-yellow-400" />
            </span>
            <span>{listing.quantity}마리</span>
          </div>
        </div>
      </div>
    </Link>
  )
}
