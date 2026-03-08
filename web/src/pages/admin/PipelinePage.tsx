import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { apiClient } from '../../api/client'
import { useAuthStore } from '../../store/authStore'

// ── 타입 정의 ─────────────────────────────────────────────

interface TranslationStat {
  locale: string
  count: number
}

interface PipelineStats {
  total_species: number
  published: number
  draft: number
  rejected: number
  pending_crawl: number
  translation_stats: TranslationStat[]
}

interface CrawlJob {
  id: number
  source_name: string
  source_url: string
  job_type: string
  status: string
  items_found: number
  items_processed: number
  items_failed: number
  error_message: string | null
  started_at: string | null
  completed_at: string | null
  created_at: string
}

// ── 상수 ──────────────────────────────────────────────────

const SUPPORTED_LOCALES = [
  'ko', 'en-US', 'en-GB', 'en-AU', 'ja',
  'zh-CN', 'zh-TW', 'de', 'fr-FR', 'fr-CA',
  'es', 'pt', 'ar', 'he',
]

// ── 헬퍼 ──────────────────────────────────────────────────

function formatDuration(startedAt: string | null, completedAt: string | null): string {
  if (!startedAt) return '-'
  const start = new Date(startedAt).getTime()
  const end = completedAt ? new Date(completedAt).getTime() : Date.now()
  const secs = Math.round((end - start) / 1000)
  if (secs < 60) return `${secs}s`
  const mins = Math.floor(secs / 60)
  const rem = secs % 60
  return `${mins}m ${rem}s`
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('ko-KR', {
    month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit',
  })
}

// ── 서브 컴포넌트 ─────────────────────────────────────────

function StatCard({
  label, value, color,
}: {
  label: string
  value: string | number
  color: 'green' | 'yellow' | 'red' | 'blue' | 'gray' | 'purple'
}) {
  const styles: Record<string, string> = {
    green:  'bg-green-50  border-green-200  text-green-700',
    yellow: 'bg-yellow-50 border-yellow-200 text-yellow-700',
    red:    'bg-red-50    border-red-200    text-red-700',
    blue:   'bg-blue-50   border-blue-200   text-blue-700',
    gray:   'bg-gray-50   border-gray-200   text-gray-700',
    purple: 'bg-purple-50 border-purple-200 text-purple-700',
  }
  return (
    <div className={`border rounded-lg p-4 ${styles[color]}`}>
      <div className="text-sm font-medium opacity-75">{label}</div>
      <div className="text-3xl font-bold mt-1">{value.toLocaleString()}</div>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    COMPLETED: 'bg-green-100 text-green-700',
    FAILED:    'bg-red-100   text-red-700',
    RUNNING:   'bg-blue-100  text-blue-700',
    PENDING:   'bg-gray-100  text-gray-600',
  }
  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium ${styles[status] ?? 'bg-gray-100 text-gray-600'}`}>
      {status}
    </span>
  )
}

function ActionButton({
  label,
  onClick,
  loading,
  color = 'blue',
}: {
  label: string
  onClick: () => void
  loading: boolean
  color?: 'blue' | 'violet' | 'teal'
}) {
  const base = 'px-4 py-2 rounded font-medium text-white text-sm transition-opacity disabled:opacity-60 flex items-center gap-2'
  const colorStyle: Record<string, string> = {
    blue:   'bg-blue-600   hover:bg-blue-700',
    violet: 'bg-violet-600 hover:bg-violet-700',
    teal:   'bg-teal-600   hover:bg-teal-700',
  }
  return (
    <button className={`${base} ${colorStyle[color]}`} onClick={onClick} disabled={loading}>
      {loading && (
        <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
      )}
      {label}
    </button>
  )
}

// ── Toast 상태 ────────────────────────────────────────────

interface Toast {
  id: number
  message: string
  type: 'success' | 'error'
}

// ── 메인 페이지 ───────────────────────────────────────────

export default function PipelinePage() {
  const { user } = useAuthStore()
  const queryClient = useQueryClient()
  const [toasts, setToasts] = useState<Toast[]>([])
  const [toastCounter, setToastCounter] = useState(0)

  const addToast = (message: string, type: 'success' | 'error') => {
    const id = toastCounter + 1
    setToastCounter(id)
    setToasts((prev) => [...prev, { id, message, type }])
    setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== id)), 3500)
  }

  // ── 데이터 조회 ─────────────────────────────────────────

  const { data: stats, isLoading: statsLoading } = useQuery<PipelineStats>({
    queryKey: ['pipeline-stats'],
    queryFn: () => apiClient.get('/admin/pipeline/stats').then((r) => r.data),
    refetchInterval: 5000,
  })

  const { data: jobsData, isLoading: jobsLoading } = useQuery<{ jobs: CrawlJob[] }>({
    queryKey: ['pipeline-jobs'],
    queryFn: () => apiClient.get('/admin/pipeline/jobs').then((r) => r.data),
    refetchInterval: 5000,
  })

  // ── 액션 뮤테이션 ───────────────────────────────────────

  const processMutation = useMutation({
    mutationFn: () => apiClient.post('/admin/pipeline/process'),
    onSuccess: () => {
      addToast('PENDING 처리 작업이 시작되었습니다.', 'success')
      queryClient.invalidateQueries({ queryKey: ['pipeline-stats'] })
      queryClient.invalidateQueries({ queryKey: ['pipeline-jobs'] })
    },
    onError: () => addToast('PENDING 처리 시작에 실패했습니다.', 'error'),
  })

  const enrichMutation = useMutation({
    mutationFn: () => apiClient.post('/admin/pipeline/enrich'),
    onSuccess: () => addToast('AI 품질 보강 작업이 시작되었습니다.', 'success'),
    onError: () => addToast('AI 품질 보강 시작에 실패했습니다.', 'error'),
  })

  const translateMutation = useMutation({
    mutationFn: () => apiClient.post('/admin/pipeline/translate'),
    onSuccess: () => addToast('AI 번역 작업이 시작되었습니다.', 'success'),
    onError: () => addToast('AI 번역 시작에 실패했습니다.', 'error'),
  })

  if (user?.role !== 'ADMIN') {
    return <div className="p-8 text-red-500">접근 권한이 없습니다.</div>
  }

  const jobs = jobsData?.jobs ?? []
  const total = stats?.total_species ?? 0
  const translationMap: Record<string, number> = {}
  stats?.translation_stats?.forEach((t) => { translationMap[t.locale] = t.count })

  return (
    <div className="p-6 max-w-7xl mx-auto">

      {/* ── 토스트 ── */}
      <div className="fixed top-4 right-4 z-50 flex flex-col gap-2">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={`px-4 py-3 rounded-lg shadow-lg text-white text-sm font-medium transition-all ${
              t.type === 'success' ? 'bg-green-600' : 'bg-red-600'
            }`}
          >
            {t.message}
          </div>
        ))}
      </div>

      {/* ── 헤더 & 네비게이션 ── */}
      <h1 className="text-2xl font-bold mb-4">파이프라인 관리</h1>
      <div className="flex flex-wrap gap-3 mb-8">
        <Link to="/admin" className="px-4 py-2 bg-gray-600 text-white rounded text-sm hover:bg-gray-700">
          대시보드
        </Link>
        <Link to="/admin/users" className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700">
          사용자 관리
        </Link>
        <Link to="/admin/pipeline" className="px-4 py-2 bg-indigo-700 text-white rounded text-sm cursor-default">
          파이프라인
        </Link>
      </div>

      {/* ── 통계 카드 ── */}
      <section className="mb-8">
        <h2 className="text-lg font-semibold mb-3 text-gray-700">어종 현황</h2>
        {statsLoading ? (
          <div className="text-gray-400">로딩 중...</div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
            <StatCard label="전체 어종" value={total} color="purple" />
            <StatCard label="Published" value={stats?.published ?? 0} color="green" />
            <StatCard label="Draft" value={stats?.draft ?? 0} color="yellow" />
            <StatCard label="Rejected" value={stats?.rejected ?? 0} color="red" />
            <StatCard label="Pending 처리" value={stats?.pending_crawl ?? 0} color="blue" />
          </div>
        )}
      </section>

      {/* ── 액션 버튼 ── */}
      <section className="mb-8">
        <h2 className="text-lg font-semibold mb-3 text-gray-700">파이프라인 액션</h2>
        <div className="flex flex-wrap gap-3">
          <ActionButton
            label="PENDING 처리"
            color="blue"
            loading={processMutation.isPending}
            onClick={() => processMutation.mutate()}
          />
          <ActionButton
            label="AI 품질 보강"
            color="violet"
            loading={enrichMutation.isPending}
            onClick={() => enrichMutation.mutate()}
          />
          <ActionButton
            label="AI 번역"
            color="teal"
            loading={translateMutation.isPending}
            onClick={() => translateMutation.mutate()}
          />
        </div>
        <p className="mt-2 text-xs text-gray-400">
          각 작업은 서버에서 비동기로 실행됩니다. 완료 후 통계가 자동 갱신됩니다 (5초마다).
        </p>
      </section>

      {/* ── 번역 현황 ── */}
      <section className="mb-8">
        <h2 className="text-lg font-semibold mb-3 text-gray-700">번역 현황 (14개 언어)</h2>
        {statsLoading ? (
          <div className="text-gray-400">로딩 중...</div>
        ) : (
          <div className="bg-white border rounded-lg p-4 space-y-2">
            {SUPPORTED_LOCALES.map((locale) => {
              const count = translationMap[locale] ?? 0
              const pct = total > 0 ? Math.min(100, Math.round((count / total) * 100)) : 0
              return (
                <div key={locale} className="flex items-center gap-3">
                  <span className="w-16 text-xs font-mono text-gray-600 text-right flex-shrink-0">
                    {locale}
                  </span>
                  <div className="flex-1 bg-gray-100 rounded-full h-3 overflow-hidden">
                    <div
                      className="h-3 rounded-full bg-indigo-500 transition-all"
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                  <span className="w-24 text-xs text-gray-500 flex-shrink-0">
                    {count.toLocaleString()} / {total.toLocaleString()} ({pct}%)
                  </span>
                </div>
              )
            })}
          </div>
        )}
      </section>

      {/* ── 최근 크롤 잡 ── */}
      <section>
        <h2 className="text-lg font-semibold mb-3 text-gray-700">최근 크롤 잡 (최대 20개)</h2>
        {jobsLoading ? (
          <div className="text-gray-400">로딩 중...</div>
        ) : jobs.length === 0 ? (
          <div className="text-gray-400 py-8 text-center border rounded-lg">크롤 잡 기록이 없습니다.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm border-collapse">
              <thead>
                <tr className="bg-gray-50">
                  <th className="border px-3 py-2 text-left">소스</th>
                  <th className="border px-3 py-2 text-center">타입</th>
                  <th className="border px-3 py-2 text-center">상태</th>
                  <th className="border px-3 py-2 text-right">수집</th>
                  <th className="border px-3 py-2 text-right">처리</th>
                  <th className="border px-3 py-2 text-right">실패</th>
                  <th className="border px-3 py-2 text-center">소요 시간</th>
                  <th className="border px-3 py-2 text-center">생성일</th>
                </tr>
              </thead>
              <tbody>
                {jobs.map((job) => (
                  <tr
                    key={job.id}
                    className={`hover:bg-gray-50 ${job.status === 'FAILED' ? 'bg-red-50' : ''}`}
                  >
                    <td className="border px-3 py-2">
                      <div className="font-medium">{job.source_name}</div>
                      {job.error_message && (
                        <div className="text-xs text-red-500 mt-0.5 truncate max-w-xs" title={job.error_message}>
                          {job.error_message}
                        </div>
                      )}
                    </td>
                    <td className="border px-3 py-2 text-center">
                      <span className="px-2 py-0.5 rounded text-xs bg-gray-100 text-gray-700 font-mono">
                        {job.job_type}
                      </span>
                    </td>
                    <td className="border px-3 py-2 text-center">
                      <StatusBadge status={job.status} />
                    </td>
                    <td className="border px-3 py-2 text-right tabular-nums">{job.items_found.toLocaleString()}</td>
                    <td className="border px-3 py-2 text-right tabular-nums">{job.items_processed.toLocaleString()}</td>
                    <td className="border px-3 py-2 text-right tabular-nums">
                      <span className={job.items_failed > 0 ? 'text-red-600 font-bold' : ''}>
                        {job.items_failed.toLocaleString()}
                      </span>
                    </td>
                    <td className="border px-3 py-2 text-center text-gray-600">
                      {formatDuration(job.started_at, job.completed_at)}
                    </td>
                    <td className="border px-3 py-2 text-center text-gray-500 text-xs">
                      {formatDate(job.created_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  )
}
