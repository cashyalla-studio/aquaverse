import { useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { ArrowLeft, Droplets, Thermometer, Ruler, Clock, AlertTriangle } from 'lucide-react'
import { fishApi } from '../api/fish'
import { clsx } from 'clsx'
import { events } from '../lib/posthog'

const CATEGORY_EMOJI: Record<string, string> = {
  fish: '🐠', reptile: '🦎', amphibian: '🐸', insect: '🐜',
  arachnid: '🕷️', bird: '🦜', mammal: '🐹',
}
const CATEGORY_NAME: Record<string, string> = {
  fish: '열대어', reptile: '파충류', amphibian: '양서류', insect: '곤충',
  arachnid: '거미류', bird: '조류', mammal: '소동물',
}
const getCategoryEmoji = (cat: string) => CATEGORY_EMOJI[cat] ?? '🐾'
const getCategoryName = (cat: string) => CATEGORY_NAME[cat] ?? cat

export default function FishDetail() {
  const { id } = useParams<{ id: string }>()
  const { t, i18n } = useTranslation()

  const { data: fish, isLoading, isError } = useQuery({
    queryKey: ['fish', id, i18n.language],
    queryFn: () => fishApi.get(Number(id), i18n.language).then((r) => r.data),
    enabled: !!id,
  })

  useEffect(() => {
    if (fish) {
      const displayName = fish.translation?.common_name || fish.primary_common_name
      events.speciesViewed(fish.id, displayName)
      document.title = `${displayName} (${fish.scientific_name}) - Finara`
      let meta = document.querySelector('meta[name="description"]') as HTMLMetaElement
      if (!meta) {
        meta = document.createElement('meta')
        meta.name = 'description'
        document.head.appendChild(meta)
      }
      meta.content = `${displayName} 사육 정보. 수온 ${fish.temp_min_c ?? '?'}~${fish.temp_max_c ?? '?'}°C, pH ${fish.ph_min ?? '?'}~${fish.ph_max ?? '?'}. Finara에서 분양하기.`
    }
  }, [fish])

  if (isLoading) return (
    <div className="animate-pulse space-y-4">
      <div className="h-64 bg-gray-200 rounded-xl" />
      <div className="h-8 bg-gray-200 rounded w-1/2" />
      <div className="grid grid-cols-2 gap-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="h-20 bg-gray-200 rounded-xl" />
        ))}
      </div>
    </div>
  )

  if (isError || !fish) return (
    <div className="text-center py-12">
      <p className="text-gray-500">{t('common.error')}</p>
      <Link to="/fish" className="text-primary-600 mt-2 inline-block">← {t('nav.encyclopedia')}</Link>
    </div>
  )

  // 번역 우선 적용
  const careNotes = fish.translation?.care_notes || fish.care_notes
  const breedingNotes = fish.translation?.breeding_notes || fish.breeding_notes
  const dietNotes = fish.translation?.diet_notes || fish.diet_notes

  return (
    <div className="max-w-3xl mx-auto">
      {/* 뒤로 */}
      <Link to="/fish" className="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-primary-600 mb-4">
        <ArrowLeft size={16} /> {t('nav.encyclopedia')}
      </Link>

      {/* 히어로 */}
      <div className="bg-white rounded-2xl shadow-sm overflow-hidden mb-6">
        <div className="h-64 bg-gradient-to-br from-blue-50 to-cyan-100 relative">
          {fish.primary_image_url ? (
            <img src={fish.primary_image_url} alt={fish.primary_common_name} className="w-full h-full object-cover" />
          ) : (
            <div className="flex items-center justify-center h-full text-8xl">🐠</div>
          )}
        </div>

        <div className="p-6">
          <h1 className="text-2xl font-bold text-gray-900">
            {fish.translation?.common_name || fish.primary_common_name}
          </h1>
          <p className="text-gray-400 italic mt-1">{fish.scientific_name}</p>
          <p className="text-sm text-gray-500 mt-1">Family: {fish.family}</p>

          {/* 빠른 배지 */}
          <div className="flex flex-wrap gap-2 mt-3">
            {fish.creature_category && fish.creature_category !== 'fish' && (
              <span className="inline-flex items-center gap-1 px-3 py-1 bg-amber-100 text-amber-800 rounded-full text-sm">
                {getCategoryEmoji(fish.creature_category)} {getCategoryName(fish.creature_category)}
              </span>
            )}
            {fish.care_level && (
              <span className={clsx('text-xs px-3 py-1 rounded-full font-medium', {
                'bg-green-100 text-green-700': fish.care_level === 'BEGINNER',
                'bg-yellow-100 text-yellow-700': fish.care_level === 'INTERMEDIATE',
                'bg-red-100 text-red-700': fish.care_level === 'EXPERT',
              })}>
                {t(`fish.care_levels.${fish.care_level}`)}
              </span>
            )}
            {fish.temperament && (
              <span className="text-xs px-3 py-1 rounded-full bg-blue-100 text-blue-700 font-medium">
                {t(`fish.temperaments.${fish.temperament}`)}
              </span>
            )}
            {fish.diet_type && (
              <span className="text-xs px-3 py-1 rounded-full bg-purple-100 text-purple-700 font-medium">
                {t(`fish.diet_types.${fish.diet_type}`)}
              </span>
            )}
          </div>
        </div>
      </div>

      {/* 수치 파라미터 그리드 */}
      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 mb-6">
        <ParamCard
          icon={<Ruler size={18} className="text-primary-500" />}
          label={t('fish.max_size')}
          value={fish.max_size_cm ? `${fish.max_size_cm} cm` : '—'}
        />
        <ParamCard
          icon={<Droplets size={18} className="text-blue-500" />}
          label={t('fish.tank_size')}
          value={fish.min_tank_size_liters ? `${fish.min_tank_size_liters} L` : '—'}
        />
        <ParamCard
          icon={<Droplets size={18} className="text-cyan-500" />}
          label={t('fish.ph_range')}
          value={(fish.ph_min && fish.ph_max) ? `${fish.ph_min} – ${fish.ph_max}` : '—'}
        />
        <ParamCard
          icon={<Thermometer size={18} className="text-orange-500" />}
          label={t('fish.temp_range')}
          value={(fish.temp_min_c && fish.temp_max_c) ? `${fish.temp_min_c}–${fish.temp_max_c}°C` : '—'}
        />
        <ParamCard
          icon={<Clock size={18} className="text-gray-500" />}
          label={t('fish.lifespan')}
          value={fish.lifespan_years ? `${fish.lifespan_years} yrs` : '—'}
        />
      </div>

      {/* 텍스트 섹션들 */}
      {careNotes && (
        <Section title={t('fish.care_notes')}>{careNotes}</Section>
      )}
      {dietNotes && (
        <Section title={t('fish.diet')}>{dietNotes}</Section>
      )}
      {breedingNotes && (
        <Section title={t('fish.breeding')}>{breedingNotes}</Section>
      )}

      {/* 추가 사육 정보 (extra_attributes) */}
      {fish.extra_attributes && (
        <div className="bg-gray-50 rounded-xl p-4 space-y-3 mb-4">
          <h3 className="font-semibold text-gray-800">추가 사육 정보</h3>
          {fish.extra_attributes.humidity_min != null && (
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">습도</span>
              <span>{fish.extra_attributes.humidity_min}–{fish.extra_attributes.humidity_max}%</span>
            </div>
          )}
          {fish.extra_attributes.uv_requirement && (
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">UV 조명</span>
              <span className="capitalize">{fish.extra_attributes.uv_requirement}</span>
            </div>
          )}
          {fish.extra_attributes.basking_temp_c != null && (
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">일광욕 온도</span>
              <span>{fish.extra_attributes.basking_temp_c}°C</span>
            </div>
          )}
          {fish.extra_attributes.colony_size_min != null && (
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">군체 크기</span>
              <span>{fish.extra_attributes.colony_size_min}–{fish.extra_attributes.colony_size_max ?? '?'} 마리</span>
            </div>
          )}
          {fish.extra_attributes.venom_level && (
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">독성</span>
              <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                fish.extra_attributes.venom_level === 'dangerous' ? 'bg-red-100 text-red-700' :
                fish.extra_attributes.venom_level === 'moderate' ? 'bg-orange-100 text-orange-700' :
                'bg-green-100 text-green-700'
              }`}>{fish.extra_attributes.venom_level}</span>
            </div>
          )}
          {fish.extra_attributes.lifespan_years_min != null && (
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">수명</span>
              <span>{fish.extra_attributes.lifespan_years_min}–{fish.extra_attributes.lifespan_years_max ?? '?'}년</span>
            </div>
          )}
          {fish.extra_attributes.legal_status_kr && fish.extra_attributes.legal_status_kr !== 'legal' && (
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">법적 지위</span>
              <span className="px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-700">
                {fish.extra_attributes.legal_status_kr === 'cites_appendix1' ? 'CITES 부속서 I' :
                 fish.extra_attributes.legal_status_kr === 'cites_appendix2' ? 'CITES 부속서 II' :
                 fish.extra_attributes.legal_status_kr === 'restricted' ? '규제 대상' :
                 fish.extra_attributes.legal_status_kr === 'banned' ? '반입 금지' : fish.extra_attributes.legal_status_kr}
              </span>
            </div>
          )}
        </div>
      )}

      {/* 저작권 표시 */}
      {fish.attribution && (
        <div className="mt-6 p-4 bg-gray-50 rounded-xl text-xs text-gray-400">
          <AlertTriangle size={12} className="inline mr-1" />
          {fish.attribution} | {fish.license}
        </div>
      )}

      {/* 분양마켓 연계 */}
      <div className="mt-6 p-5 bg-primary-50 rounded-xl flex items-center justify-between">
        <div>
          <p className="font-semibold text-primary-800">
            Looking for {fish.translation?.common_name || fish.primary_common_name}?
          </p>
          <p className="text-sm text-primary-600 mt-0.5">Check the adoption market</p>
        </div>
        <Link
          to={`/marketplace?fish_id=${fish.id}`}
          className="bg-primary-500 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary-600"
        >
          {t('nav.marketplace')} →
        </Link>
      </div>
    </div>
  )
}

function ParamCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div className="bg-white rounded-xl shadow-sm p-4 flex items-center gap-3">
      {icon}
      <div>
        <p className="text-xs text-gray-400">{label}</p>
        <p className="font-semibold text-gray-900 text-sm">{value}</p>
      </div>
    </div>
  )
}

function Section({ title, children }: { title: string; children: string }) {
  return (
    <div className="bg-white rounded-xl shadow-sm p-5 mb-4">
      <h2 className="font-semibold text-gray-800 mb-2">{title}</h2>
      <p className="text-gray-600 text-sm leading-relaxed whitespace-pre-line">{children}</p>
    </div>
  )
}
