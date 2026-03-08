import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import {
  Calendar,
  CheckCircle2,
  Flame,
  Plus,
  Droplets,
  Utensils,
  Filter,
  Pill,
  Star,
  Trophy,
  X,
  Clock,
} from 'lucide-react'
import { careApi, type CareSchedule } from '../api/care'
import { useAuthStore } from '../store/authStore'
import { clsx } from 'clsx'

// ── 타입별 설정 ──────────────────────────────────────────────────────────────
const TYPE_CONFIG = {
  feeding: {
    label: '먹이',
    icon: Utensils,
    color: 'text-orange-500',
    bg: 'bg-orange-50',
    border: 'border-orange-200',
  },
  water_change: {
    label: '수질교환',
    icon: Droplets,
    color: 'text-blue-500',
    bg: 'bg-blue-50',
    border: 'border-blue-200',
  },
  filter_clean: {
    label: '필터청소',
    icon: Filter,
    color: 'text-purple-500',
    bg: 'bg-purple-50',
    border: 'border-purple-200',
  },
  medication: {
    label: '투약',
    icon: Pill,
    color: 'text-red-500',
    bg: 'bg-red-50',
    border: 'border-red-200',
  },
  custom: {
    label: '기타',
    icon: Star,
    color: 'text-gray-500',
    bg: 'bg-gray-50',
    border: 'border-gray-200',
  },
} as const

const FREQUENCY_LABELS: Record<string, string> = {
  daily: '매일',
  weekly: '매주',
  biweekly: '격주',
  monthly: '매월',
  custom: '직접 설정',
}

// ── 유틸: 날짜 포맷 ──────────────────────────────────────────────────────────
function formatDate(iso: string) {
  const d = new Date(iso)
  return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function formatDateLabel(iso: string) {
  const d = new Date(iso)
  const today = new Date()
  const diff = Math.floor((d.getTime() - today.setHours(0, 0, 0, 0)) / 86400000)
  if (diff === 0) return '오늘'
  if (diff === 1) return '내일'
  if (diff === 2) return '모레'
  return `${d.getMonth() + 1}월 ${d.getDate()}일`
}

function getDayKey(iso: string) {
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// ── 스트릭 카드 ───────────────────────────────────────────────────────────────
function StreakCard() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

  const { data: streak, isLoading } = useQuery({
    queryKey: ['care-streak'],
    queryFn: () => careApi.getStreak().then((r) => r.data),
    enabled: isAuthenticated,
  })

  if (!isAuthenticated) return null

  return (
    <div className="bg-gradient-to-r from-orange-400 to-red-500 rounded-2xl p-5 text-white">
      <div className="flex items-center justify-between">
        {/* 왼쪽: 현재 스트릭 */}
        <div className="flex items-center gap-3">
          <div className="text-4xl">🔥</div>
          <div>
            <div className="flex items-baseline gap-1">
              <span className="text-4xl font-black">
                {isLoading ? '—' : (streak?.current_streak ?? 0)}
              </span>
              <span className="text-lg font-semibold opacity-90">일</span>
            </div>
            <p className="text-sm font-medium opacity-80 mt-0.5">연속 케어 중</p>
          </div>
        </div>

        {/* 오른쪽: 최장 기록 */}
        <div className="text-right">
          <div className="flex items-center gap-1.5 justify-end mb-1">
            <Trophy size={14} className="opacity-80" />
            <span className="text-xs font-medium opacity-80">최장 기록</span>
          </div>
          <div className="text-2xl font-bold">
            {isLoading ? '—' : (streak?.longest_streak ?? 0)}
            <span className="text-sm font-normal opacity-80">일</span>
          </div>
          {streak?.last_care_date && (
            <p className="text-xs opacity-60 mt-0.5">
              마지막: {new Date(streak.last_care_date).toLocaleDateString('ko-KR')}
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

// ── 일정 카드 ─────────────────────────────────────────────────────────────────
function ScheduleCard({
  schedule,
  onComplete,
  completing,
  done,
}: {
  schedule: CareSchedule
  onComplete: (id: number) => void
  completing: boolean
  done: boolean
}) {
  const cfg = TYPE_CONFIG[schedule.schedule_type] ?? TYPE_CONFIG.custom
  const Icon = cfg.icon

  return (
    <div
      className={clsx(
        'card p-4 flex items-center gap-3 transition-all duration-300',
        done && 'opacity-50',
      )}
    >
      {/* 타입 아이콘 */}
      <div
        className={clsx(
          'w-11 h-11 rounded-full flex items-center justify-center flex-shrink-0',
          cfg.bg,
        )}
      >
        <Icon size={20} className={cfg.color} />
      </div>

      {/* 중앙: 제목 + 부제 */}
      <div className="flex-1 min-w-0">
        <p
          className={clsx(
            'font-semibold text-gray-900 text-sm truncate',
            done && 'line-through text-gray-400',
          )}
        >
          {schedule.title}
        </p>
        <div className="flex items-center gap-2 mt-0.5">
          <span className="text-xs text-gray-400">{cfg.label}</span>
          {schedule.next_due_at && (
            <>
              <span className="text-gray-200">·</span>
              <span className="flex items-center gap-0.5 text-xs text-gray-400">
                <Clock size={10} />
                {formatDate(schedule.next_due_at)}
              </span>
            </>
          )}
          <span className={clsx('badge text-xs', cfg.bg, cfg.color, `border ${cfg.border}`)}>
            {FREQUENCY_LABELS[schedule.frequency] ?? schedule.frequency}
          </span>
        </div>
        {schedule.description && (
          <p className="text-xs text-gray-400 truncate mt-0.5">{schedule.description}</p>
        )}
      </div>

      {/* 오른쪽: 완료 버튼 */}
      {done ? (
        <CheckCircle2 size={24} className="text-emerald-400 flex-shrink-0" />
      ) : (
        <button
          onClick={() => onComplete(schedule.id)}
          disabled={completing}
          className={clsx(
            'flex-shrink-0 w-9 h-9 rounded-full border-2 border-emerald-400 flex items-center justify-center',
            'hover:bg-emerald-50 active:scale-95 transition-all duration-150',
            completing && 'opacity-50 cursor-not-allowed',
          )}
          aria-label="완료"
        >
          {completing ? (
            <div className="w-4 h-4 border-2 border-emerald-400 border-t-transparent rounded-full animate-spin" />
          ) : (
            <CheckCircle2 size={18} className="text-emerald-400" />
          )}
        </button>
      )}
    </div>
  )
}

// ── 일정 추가 모달 ────────────────────────────────────────────────────────────
interface AddModalProps {
  open: boolean
  onClose: () => void
  tankId: number
}

function AddScheduleModal({ open, onClose, tankId }: AddModalProps) {
  const queryClient = useQueryClient()
  const [form, setForm] = useState({
    title: '',
    schedule_type: 'feeding' as CareSchedule['schedule_type'],
    frequency: 'daily' as CareSchedule['frequency'],
    interval_days: 1,
    next_due_at: new Date().toISOString().slice(0, 16),
    description: '',
  })

  const { mutate, isPending } = useMutation({
    mutationFn: () =>
      careApi.createSchedule(tankId, {
        ...form,
        next_due_at: new Date(form.next_due_at).toISOString(),
        interval_days: form.frequency === 'custom' ? form.interval_days : undefined,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['care-schedules', tankId] })
      queryClient.invalidateQueries({ queryKey: ['care-today'] })
      onClose()
      setForm({
        title: '',
        schedule_type: 'feeding',
        frequency: 'daily',
        interval_days: 1,
        next_due_at: new Date().toISOString().slice(0, 16),
        description: '',
      })
    },
  })

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center">
      {/* 오버레이 */}
      <div
        className="absolute inset-0 bg-black/40 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* 모달 */}
      <div className="relative w-full sm:max-w-md bg-white rounded-t-3xl sm:rounded-2xl p-6 shadow-xl">
        {/* 헤더 */}
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-lg font-bold text-gray-900">케어 일정 추가</h2>
          <button
            onClick={onClose}
            className="p-2 rounded-xl text-gray-400 hover:bg-gray-100 transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        <div className="space-y-4">
          {/* 제목 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">제목</label>
            <input
              type="text"
              placeholder="예: 아침 먹이주기"
              value={form.title}
              onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
              className="input"
            />
          </div>

          {/* 수조 (MVP: 하드코딩) */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">수조</label>
            <div className="input bg-gray-50 text-gray-500 cursor-not-allowed">
              내 수조 #{tankId} (MVP)
            </div>
          </div>

          {/* 종류 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">케어 종류</label>
            <div className="grid grid-cols-3 gap-2">
              {(Object.keys(TYPE_CONFIG) as Array<keyof typeof TYPE_CONFIG>).map((type) => {
                const cfg = TYPE_CONFIG[type]
                const Icon = cfg.icon
                return (
                  <button
                    key={type}
                    type="button"
                    onClick={() => setForm((f) => ({ ...f, schedule_type: type }))}
                    className={clsx(
                      'flex flex-col items-center gap-1.5 p-3 rounded-xl border-2 text-xs font-medium transition-all',
                      form.schedule_type === type
                        ? `${cfg.bg} ${cfg.color} border-current`
                        : 'bg-white text-gray-500 border-gray-200 hover:border-gray-300',
                    )}
                  >
                    <Icon size={18} />
                    {cfg.label}
                  </button>
                )
              })}
            </div>
          </div>

          {/* 주기 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">반복 주기</label>
            <select
              value={form.frequency}
              onChange={(e) =>
                setForm((f) => ({ ...f, frequency: e.target.value as CareSchedule['frequency'] }))
              }
              className="input"
            >
              {Object.entries(FREQUENCY_LABELS).map(([val, label]) => (
                <option key={val} value={val}>
                  {label}
                </option>
              ))}
            </select>
            {form.frequency === 'custom' && (
              <div className="mt-2 flex items-center gap-2">
                <input
                  type="number"
                  min={1}
                  max={365}
                  value={form.interval_days}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, interval_days: Number(e.target.value) }))
                  }
                  className="input w-24"
                />
                <span className="text-sm text-gray-500">일마다</span>
              </div>
            )}
          </div>

          {/* 첫 시작 날짜 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">첫 케어 날짜</label>
            <input
              type="datetime-local"
              value={form.next_due_at}
              onChange={(e) => setForm((f) => ({ ...f, next_due_at: e.target.value }))}
              className="input"
            />
          </div>

          {/* 메모 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">메모 (선택)</label>
            <textarea
              placeholder="추가 설명을 입력하세요..."
              value={form.description}
              onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              rows={2}
              className="input resize-none"
            />
          </div>
        </div>

        {/* 액션 버튼 */}
        <div className="flex gap-3 mt-6">
          <button onClick={onClose} className="btn-secondary flex-1">
            취소
          </button>
          <button
            onClick={() => mutate()}
            disabled={!form.title.trim() || isPending}
            className="btn-primary flex-1 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isPending ? '저장 중...' : '저장'}
          </button>
        </div>
      </div>
    </div>
  )
}

// ── 탭별 콘텐츠: 오늘 ─────────────────────────────────────────────────────────
function TodayTab({ tankId }: { tankId: number }) {
  const queryClient = useQueryClient()
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const [completedIds, setCompletedIds] = useState<Set<number>>(new Set())
  const [completingId, setCompletingId] = useState<number | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['care-today'],
    queryFn: () => careApi.getTodayTasks().then((r) => r.data),
    enabled: isAuthenticated,
  })

  const { mutate: complete } = useMutation({
    mutationFn: (id: number) => careApi.completeSchedule(id),
    onMutate: (id) => setCompletingId(id),
    onSuccess: (_, id) => {
      setCompletedIds((prev) => new Set([...prev, id]))
      setCompletingId(null)
      queryClient.invalidateQueries({ queryKey: ['care-streak'] })
      queryClient.invalidateQueries({ queryKey: ['care-today'] })
      queryClient.invalidateQueries({ queryKey: ['care-schedules', tankId] })
    },
    onError: () => setCompletingId(null),
  })

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="card p-4 flex items-center gap-3 animate-pulse">
            <div className="w-11 h-11 rounded-full bg-gray-100 flex-shrink-0" />
            <div className="flex-1 space-y-2">
              <div className="h-4 bg-gray-100 rounded-lg w-1/2" />
              <div className="h-3 bg-gray-100 rounded-lg w-1/3" />
            </div>
          </div>
        ))}
      </div>
    )
  }

  const tasks = data?.tasks ?? []
  const pending = tasks.filter((t) => !completedIds.has(t.id))
  const done = tasks.filter((t) => completedIds.has(t.id))

  if (tasks.length === 0) {
    return (
      <div className="text-center py-16">
        <p className="text-5xl mb-3">🎉</p>
        <p className="text-gray-700 font-semibold">오늘 케어 일정이 없어요!</p>
        <p className="text-gray-400 text-sm mt-1">모든 일정을 마쳤거나 아직 등록이 없습니다.</p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {/* 남은 일정 */}
      {pending.map((s) => (
        <ScheduleCard
          key={s.id}
          schedule={s}
          onComplete={complete}
          completing={completingId === s.id}
          done={false}
        />
      ))}

      {/* 완료된 일정 */}
      {done.length > 0 && (
        <>
          <div className="flex items-center gap-2 pt-2">
            <div className="flex-1 h-px bg-gray-100" />
            <span className="text-xs text-gray-400 font-medium">완료 {done.length}개</span>
            <div className="flex-1 h-px bg-gray-100" />
          </div>
          {done.map((s) => (
            <ScheduleCard
              key={s.id}
              schedule={s}
              onComplete={complete}
              completing={false}
              done={true}
            />
          ))}
        </>
      )}
    </div>
  )
}

// ── 탭별 콘텐츠: 예정 ─────────────────────────────────────────────────────────
function UpcomingTab({ tankId }: { tankId: number }) {
  const { data, isLoading } = useQuery({
    queryKey: ['care-schedules', tankId],
    queryFn: () => careApi.listSchedules(tankId).then((r) => r.data),
  })

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="card p-4 animate-pulse h-16" />
        ))}
      </div>
    )
  }

  const schedules = (data?.schedules ?? []).filter((s) => s.is_active)

  if (schedules.length === 0) {
    return (
      <div className="text-center py-16">
        <p className="text-5xl mb-3">📅</p>
        <p className="text-gray-700 font-semibold">예정된 일정이 없어요</p>
        <p className="text-gray-400 text-sm mt-1">+ 버튼으로 케어 일정을 추가해 보세요.</p>
      </div>
    )
  }

  // 날짜별 그룹핑
  const grouped = schedules.reduce<Record<string, CareSchedule[]>>((acc, s) => {
    const key = getDayKey(s.next_due_at)
    if (!acc[key]) acc[key] = []
    acc[key].push(s)
    return acc
  }, {})

  const sortedKeys = Object.keys(grouped).sort()

  return (
    <div className="space-y-5">
      {sortedKeys.map((key) => (
        <div key={key}>
          <div className="flex items-center gap-2 mb-2">
            <Calendar size={14} className="text-gray-400" />
            <span className="text-sm font-semibold text-gray-600">
              {formatDateLabel(grouped[key][0].next_due_at)}
            </span>
            <span className="badge badge-blue">{grouped[key].length}개</span>
          </div>
          <div className="space-y-2">
            {grouped[key].map((s) => {
              const cfg = TYPE_CONFIG[s.schedule_type] ?? TYPE_CONFIG.custom
              const Icon = cfg.icon
              return (
                <div key={s.id} className="card p-3.5 flex items-center gap-3">
                  <div
                    className={clsx(
                      'w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0',
                      cfg.bg,
                    )}
                  >
                    <Icon size={16} className={cfg.color} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-gray-900 text-sm truncate">{s.title}</p>
                    <p className="text-xs text-gray-400 mt-0.5">
                      {cfg.label} · {formatDate(s.next_due_at)}
                    </p>
                  </div>
                  <span className="text-xs text-gray-400 flex-shrink-0">
                    {FREQUENCY_LABELS[s.frequency]}
                  </span>
                </div>
              )
            })}
          </div>
        </div>
      ))}
    </div>
  )
}

// ── 탭별 콘텐츠: 완료 (더미 안내) ────────────────────────────────────────────
function DoneTab() {
  return (
    <div className="text-center py-16">
      <p className="text-5xl mb-3">✅</p>
      <p className="text-gray-700 font-semibold">완료된 케어 기록</p>
      <p className="text-gray-400 text-sm mt-1 max-w-xs mx-auto">
        완료된 케어 로그는 서버에서 care_logs 테이블로 관리됩니다.
        <br />로그 조회 API 연동 예정입니다.
      </p>
    </div>
  )
}

// ── 메인 페이지 ───────────────────────────────────────────────────────────────
type TabKey = 'today' | 'upcoming' | 'done'

const TABS: { key: TabKey; label: string; icon: typeof Calendar }[] = [
  { key: 'today', label: '오늘', icon: CheckCircle2 },
  { key: 'upcoming', label: '예정', icon: Calendar },
  { key: 'done', label: '완료', icon: Trophy },
]

// MVP: 단일 수조 하드코딩
const DEFAULT_TANK_ID = 1

export default function CareHub() {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState<TabKey>('today')
  const [showModal, setShowModal] = useState(false)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

  return (
    <div className="space-y-5">
      {/* ── 헤더 ── */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-black text-gray-900">Life Care Hub</h1>
          <p className="text-sm text-gray-400 mt-0.5">수조 케어 일정을 관리하세요</p>
        </div>
        {isAuthenticated && (
          <button
            onClick={() => setShowModal(true)}
            className="btn-primary flex items-center gap-1.5"
          >
            <Plus size={16} />
            일정 추가
          </button>
        )}
      </div>

      {/* ── 스트릭 카드 ── */}
      <StreakCard />

      {/* ── 미인증 안내 ── */}
      {!isAuthenticated && (
        <div className="card p-6 text-center">
          <p className="text-3xl mb-2">🔒</p>
          <p className="font-semibold text-gray-700">로그인이 필요한 기능입니다</p>
          <p className="text-sm text-gray-400 mt-1 mb-4">
            케어 일정을 관리하고 스트릭을 쌓으려면 로그인하세요.
          </p>
          <a href="/login" className="btn-primary inline-block">
            로그인
          </a>
        </div>
      )}

      {/* ── 탭 + 콘텐츠 ── */}
      {isAuthenticated && (
        <>
          {/* 탭 바 */}
          <div className="flex gap-1 bg-gray-100 p-1 rounded-xl">
            {TABS.map(({ key, label, icon: Icon }) => (
              <button
                key={key}
                onClick={() => setActiveTab(key)}
                className={clsx(
                  'flex-1 flex items-center justify-center gap-1.5 py-2 rounded-lg text-sm font-semibold transition-all',
                  activeTab === key
                    ? 'bg-white text-primary-600 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700',
                )}
              >
                <Icon size={15} />
                {label}
              </button>
            ))}
          </div>

          {/* 탭 콘텐츠 */}
          {activeTab === 'today' && <TodayTab tankId={DEFAULT_TANK_ID} />}
          {activeTab === 'upcoming' && <UpcomingTab tankId={DEFAULT_TANK_ID} />}
          {activeTab === 'done' && <DoneTab />}
        </>
      )}

      {/* ── 일정 추가 모달 ── */}
      <AddScheduleModal
        open={showModal}
        onClose={() => setShowModal(false)}
        tankId={DEFAULT_TANK_ID}
      />
    </div>
  )
}
