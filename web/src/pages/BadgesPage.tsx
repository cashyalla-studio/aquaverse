import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { Trophy, Lock, Medal, Swords, CheckCircle2, Clock } from 'lucide-react'
import { clsx } from 'clsx'
import { useAuthStore } from '../store/authStore'
import { badgesApi, type BadgeDefinition, type UserBadge, type Challenge } from '../api/badges'

// ── 카테고리 메타 ─────────────────────────────────────────────────
const CATEGORY_META: Record<
  string,
  { label: string; icon: string; color: string }
> = {
  care:       { label: '케어',     icon: '💧', color: 'text-cyan-600' },
  community:  { label: '커뮤니티', icon: '💬', color: 'text-violet-600' },
  market:     { label: '마켓',     icon: '🛒', color: 'text-amber-600' },
  collection: { label: '컬렉션',   icon: '🐠', color: 'text-emerald-600' },
  special:    { label: '스페셜',   icon: '✨', color: 'text-rose-600' },
}

const CATEGORY_ORDER: Array<BadgeDefinition['category']> = [
  'care', 'community', 'market', 'collection', 'special',
]

// ── 날짜 포맷 헬퍼 ────────────────────────────────────────────────
function formatDate(iso: string) {
  const d = new Date(iso)
  return `${d.getFullYear()}.${String(d.getMonth() + 1).padStart(2, '0')}.${String(d.getDate()).padStart(2, '0')}`
}

function daysLeft(iso: string) {
  const diff = new Date(iso).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

// ── 탭 타입 ───────────────────────────────────────────────────────
type Tab = 'gallery' | 'challenges'

// ── 로그인 안내 ───────────────────────────────────────────────────
function LoginPrompt() {
  return (
    <div className="card p-10 text-center space-y-4">
      <span className="text-5xl">🔐</span>
      <p className="text-lg font-semibold text-gray-700">로그인 후 내 뱃지를 확인하세요</p>
      <p className="text-sm text-gray-400">활동을 통해 뱃지를 모으고 챌린지에 참가해 보세요!</p>
      <Link to="/login" className="btn-primary inline-block mt-2">
        로그인하기
      </Link>
    </div>
  )
}

// ── 뱃지 카드 ─────────────────────────────────────────────────────
interface BadgeCardProps {
  def: BadgeDefinition
  earned?: UserBadge
}

function BadgeCard({ def, earned }: BadgeCardProps) {
  return (
    <div
      className={clsx(
        'card p-4 text-center flex flex-col items-center gap-2 transition-all duration-200',
        earned
          ? 'hover:shadow-md hover:-translate-y-0.5'
          : 'opacity-40 grayscale',
      )}
    >
      <div className="relative">
        <span className="text-4xl leading-none" role="img" aria-label={def.name}>
          {def.icon_emoji}
        </span>
        {!earned && (
          <Lock
            size={14}
            className="absolute -bottom-1 -right-1 text-gray-400 bg-white rounded-full p-0.5"
          />
        )}
      </div>
      <p className="font-semibold text-sm text-gray-900 leading-tight">{def.name}</p>
      {earned ? (
        <p className="text-xs text-gray-400">{formatDate(earned.earned_at)} 획득</p>
      ) : (
        <p className="text-xs text-gray-400">조건: {def.condition_value}개</p>
      )}
    </div>
  )
}

// ── 챌린지 카드 ───────────────────────────────────────────────────
interface ChallengeCardProps {
  challenge: Challenge
  isAuthenticated: boolean
}

function ChallengeCard({ challenge, isAuthenticated }: ChallengeCardProps) {
  const qc = useQueryClient()

  const { data: progressData } = useQuery({
    queryKey: ['challenge-progress', challenge.id],
    queryFn: () => badgesApi.getChallengeProgress(challenge.id).then((r) => r.data),
    enabled: isAuthenticated,
    staleTime: 60 * 1000,
  })

  const joinMutation = useMutation({
    mutationFn: () => badgesApi.joinChallenge(challenge.id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['challenge-progress', challenge.id] })
    },
  })

  const progress = progressData?.progress ?? 0
  const joined = !!progressData
  const completed = progressData?.completed ?? false
  const remaining = daysLeft(challenge.ends_at)

  return (
    <div className="card p-5 space-y-4">
      {/* 헤더 */}
      <div className="flex items-start justify-between gap-3">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            {challenge.badge_code && (
              <span className="badge badge-purple">
                <Medal size={10} />
                뱃지 보상
              </span>
            )}
            {completed && (
              <span className="badge badge-green">
                <CheckCircle2 size={10} />
                완료
              </span>
            )}
          </div>
          <h3 className="font-bold text-gray-900 leading-tight">{challenge.title}</h3>
          <p className="text-sm text-gray-500 mt-0.5 line-clamp-2">{challenge.description}</p>
        </div>
        <div className="flex flex-col items-end gap-1 flex-shrink-0">
          <span className="flex items-center gap-1 text-xs text-gray-400">
            <Clock size={12} />
            {remaining > 0 ? `${remaining}일 남음` : '종료'}
          </span>
          <span className="text-xs text-gray-300">{formatDate(challenge.ends_at)}</span>
        </div>
      </div>

      {/* 진행률 바 */}
      {joined && (
        <div className="space-y-1.5">
          <div className="flex justify-between text-xs text-gray-500">
            <span>진행률</span>
            <span>
              {Math.min(progress, challenge.condition_value)} / {challenge.condition_value}
            </span>
          </div>
          <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
            <div
              className={clsx(
                'h-2 rounded-full transition-all duration-500',
                completed ? 'bg-emerald-500' : 'bg-primary-500',
              )}
              style={{
                width: `${Math.min(100, (progress / challenge.condition_value) * 100)}%`,
              }}
            />
          </div>
        </div>
      )}

      {/* 참가 버튼 */}
      {isAuthenticated && (
        <div className="flex justify-end">
          {completed ? (
            <span className="flex items-center gap-1.5 text-sm font-semibold text-emerald-600">
              <CheckCircle2 size={16} />
              완료!
            </span>
          ) : joined ? (
            <span className="flex items-center gap-1.5 text-sm font-medium text-primary-600">
              <Swords size={14} />
              진행 중
            </span>
          ) : (
            <button
              onClick={() => joinMutation.mutate()}
              disabled={joinMutation.isPending || remaining === 0}
              className="btn-primary py-2 disabled:opacity-50"
            >
              {joinMutation.isPending ? '참가 중...' : '참가하기'}
            </button>
          )}
        </div>
      )}
    </div>
  )
}

// ── 뱃지 갤러리 탭 ────────────────────────────────────────────────
interface GalleryTabProps {
  allBadges: BadgeDefinition[]
  myBadges: UserBadge[]
  isAuthenticated: boolean
}

function GalleryTab({ allBadges, myBadges, isAuthenticated }: GalleryTabProps) {
  const earnedMap = new Map(myBadges.map((b) => [b.badge_code, b]))

  return (
    <div className="space-y-8">
      {!isAuthenticated && (
        <div className="card p-4 flex items-center gap-3 border-l-4 border-primary-400 bg-primary-50">
          <span className="text-xl">💡</span>
          <p className="text-sm text-primary-700">
            로그인하면 획득한 뱃지와 잠긴 뱃지를 함께 확인할 수 있어요.
          </p>
        </div>
      )}

      {CATEGORY_ORDER.map((cat) => {
        const meta = CATEGORY_META[cat]
        const items = allBadges.filter((b) => b.category === cat)
        if (items.length === 0) return null

        return (
          <section key={cat} className="space-y-4">
            {/* 카테고리 헤더 */}
            <div className="flex items-center gap-2">
              <span className="text-xl leading-none">{meta.icon}</span>
              <h2 className={clsx('section-title', meta.color)}>{meta.label}</h2>
              <span className="text-sm text-gray-400 ml-1">
                {isAuthenticated
                  ? `${items.filter((b) => earnedMap.has(b.code)).length} / ${items.length}`
                  : `${items.length}개`}
              </span>
            </div>

            {/* 뱃지 그리드 */}
            <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-8 gap-3">
              {items.map((def) => (
                <BadgeCard
                  key={def.code}
                  def={def}
                  earned={isAuthenticated ? earnedMap.get(def.code) : undefined}
                />
              ))}
            </div>
          </section>
        )
      })}
    </div>
  )
}

// ── 메인 페이지 ───────────────────────────────────────────────────
export default function BadgesPage() {
  const [activeTab, setActiveTab] = useState<Tab>('gallery')
  const { isAuthenticated } = useAuthStore()

  const { data: allBadges = [], isLoading: loadingBadges } = useQuery({
    queryKey: ['badges'],
    queryFn: () => badgesApi.listBadges().then((r) => r.data),
    staleTime: 10 * 60 * 1000,
  })

  const { data: myBadges = [], isLoading: loadingMine } = useQuery({
    queryKey: ['my-badges'],
    queryFn: () => badgesApi.getMyBadges().then((r) => r.data),
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000,
  })

  const { data: challenges = [], isLoading: loadingChallenges } = useQuery({
    queryKey: ['challenges'],
    queryFn: () => badgesApi.listChallenges().then((r) => r.data),
    staleTime: 5 * 60 * 1000,
  })

  const isLoading = loadingBadges || loadingMine || loadingChallenges

  const TABS: { id: Tab; label: string; count?: number }[] = [
    { id: 'gallery',    label: '뱃지 갤러리',    count: allBadges.length },
    { id: 'challenges', label: '진행 중 챌린지', count: challenges.length },
  ]

  return (
    <div className="space-y-6">
      {/* ── 히어로 헤더 ── */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-purple-600 via-violet-600 to-indigo-600 p-6 sm:p-8 text-white">
        {/* 배경 패턴 */}
        <div
          className="absolute inset-0 opacity-10"
          style={{
            backgroundImage:
              'radial-gradient(circle at 15% 60%, white 1px, transparent 1px), radial-gradient(circle at 85% 25%, white 1px, transparent 1px)',
            backgroundSize: '36px 36px',
          }}
        />
        <div className="relative">
          <div className="flex items-center gap-2 mb-1">
            <Trophy size={16} className="text-violet-200" />
            <span className="text-violet-200 text-sm font-medium">도전과 보상</span>
          </div>
          <h1 className="text-3xl font-black tracking-tight mb-2">나의 뱃지 컬렉션</h1>
          <p className="text-violet-100 text-sm">
            {isAuthenticated
              ? myBadges.length > 0
                ? `${myBadges.length}개의 뱃지를 획득했어요!`
                : '첫 번째 뱃지를 획득해 보세요'
              : '활동하면서 다양한 뱃지를 모아보세요'}
          </p>

          {/* 획득 통계 (로그인 시) */}
          {isAuthenticated && myBadges.length > 0 && (
            <div className="flex flex-wrap gap-3 mt-5">
              {CATEGORY_ORDER.map((cat) => {
                const meta = CATEGORY_META[cat]
                const count = myBadges.filter((b) => b.category === cat).length
                if (count === 0) return null
                return (
                  <div
                    key={cat}
                    className="flex items-center gap-1.5 bg-white/15 rounded-xl px-3 py-1.5 text-sm"
                  >
                    <span>{meta.icon}</span>
                    <span className="font-semibold">{count}</span>
                    <span className="text-white/70">{meta.label}</span>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>

      {/* ── 비로그인: 내 뱃지 없음 안내 ── */}
      {!isAuthenticated && activeTab === 'gallery' && (
        <div className="hidden">
          {/* GalleryTab 내부에서 처리 */}
        </div>
      )}

      {/* ── 탭 ── */}
      <div className="flex gap-1 border-b border-gray-100">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={clsx(
              'flex items-center gap-1.5 px-4 py-2.5 text-sm font-medium rounded-t-xl transition-colors',
              activeTab === tab.id
                ? 'text-primary-600 border-b-2 border-primary-500 -mb-px bg-white'
                : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50',
            )}
          >
            {tab.label}
            {tab.count !== undefined && tab.count > 0 && (
              <span
                className={clsx(
                  'text-xs px-1.5 py-0.5 rounded-full',
                  activeTab === tab.id
                    ? 'bg-primary-100 text-primary-600'
                    : 'bg-gray-100 text-gray-400',
                )}
              >
                {tab.count}
              </span>
            )}
          </button>
        ))}
      </div>

      {/* ── 로딩 ── */}
      {isLoading && (
        <div className="space-y-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="card p-6 animate-pulse space-y-3">
              <div className="h-5 bg-gray-100 rounded-lg w-1/4" />
              <div className="grid grid-cols-4 sm:grid-cols-6 gap-3">
                {Array.from({ length: 6 }).map((_, j) => (
                  <div key={j} className="h-24 bg-gray-100 rounded-2xl" />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ── 탭 콘텐츠 ── */}
      {!isLoading && (
        <>
          {activeTab === 'gallery' && (
            <GalleryTab
              allBadges={allBadges}
              myBadges={myBadges}
              isAuthenticated={isAuthenticated}
            />
          )}

          {activeTab === 'challenges' && (
            <div className="space-y-4">
              {!isAuthenticated && <LoginPrompt />}

              {challenges.length === 0 ? (
                <div className="text-center py-16">
                  <span className="text-5xl">🏆</span>
                  <p className="text-gray-500 font-medium mt-3">진행 중인 챌린지가 없어요</p>
                  <p className="text-gray-400 text-sm mt-1">곧 새로운 챌린지가 열릴 예정이에요</p>
                </div>
              ) : (
                <div className="grid gap-4 sm:grid-cols-2">
                  {challenges.map((challenge) => (
                    <ChallengeCard
                      key={challenge.id}
                      challenge={challenge}
                      isAuthenticated={isAuthenticated}
                    />
                  ))}
                </div>
              )}
            </div>
          )}
        </>
      )}
    </div>
  )
}
