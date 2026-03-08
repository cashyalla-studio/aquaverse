import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { CheckCircle2, ChevronRight, Droplets, Filter, Pill, Star, Utensils } from 'lucide-react'
import { Link } from 'react-router-dom'
import { careApi, type CareSchedule } from '../api/care'
import { useAuthStore } from '../store/authStore'
import { clsx } from 'clsx'
import { useState } from 'react'

// 타입별 아이콘 매핑 (CareHub.tsx TYPE_CONFIG와 동일 구조, 작은 크기용)
const TYPE_ICON: Record<CareSchedule['schedule_type'], React.ElementType> = {
  feeding: Utensils,
  water_change: Droplets,
  filter_clean: Filter,
  medication: Pill,
  custom: Star,
}

const TYPE_COLOR: Record<CareSchedule['schedule_type'], string> = {
  feeding: 'text-orange-500 bg-orange-50',
  water_change: 'text-blue-500 bg-blue-50',
  filter_clean: 'text-purple-500 bg-purple-50',
  medication: 'text-red-500 bg-red-50',
  custom: 'text-gray-500 bg-gray-50',
}

export default function CareToday() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const queryClient = useQueryClient()
  const [completedIds, setCompletedIds] = useState<Set<number>>(new Set())

  const { data, isLoading } = useQuery({
    queryKey: ['care-today'],
    queryFn: () => careApi.getTodayTasks().then((r) => r.data),
    // 로그인한 경우에만 요청
    enabled: isAuthenticated,
  })

  const { mutate: complete, variables: completingId } = useMutation({
    mutationFn: (id: number) => careApi.completeSchedule(id),
    onSuccess: (_, id) => {
      setCompletedIds((prev) => new Set([...prev, id]))
      queryClient.invalidateQueries({ queryKey: ['care-today'] })
      queryClient.invalidateQueries({ queryKey: ['care-streak'] })
    },
  })

  // 미인증 → 렌더링 없음
  if (!isAuthenticated) return null

  // 로딩 중 → 스켈레톤
  if (isLoading) {
    return (
      <div className="card p-4 animate-pulse">
        <div className="h-4 bg-gray-100 rounded-lg w-1/3 mb-3" />
        <div className="space-y-2">
          {[1, 2].map((i) => (
            <div key={i} className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-gray-100 flex-shrink-0" />
              <div className="flex-1 h-4 bg-gray-100 rounded-lg" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  const tasks = data?.tasks ?? []

  // 완료 처리되지 않은 남은 작업
  const remaining = tasks.filter((t) => !completedIds.has(t.id))

  // 남은 일정 없으면 렌더링 없음
  if (remaining.length === 0) return null

  return (
    <div className="card overflow-hidden">
      {/* 헤더 */}
      <div className="bg-gradient-to-r from-emerald-500 to-teal-500 px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-lg">🌿</span>
          <div>
            <p className="text-white font-bold text-sm">오늘 케어 {remaining.length}개 남았어요</p>
            <p className="text-emerald-100 text-xs">지금 바로 처리해 보세요</p>
          </div>
        </div>
        <Link
          to="/care"
          className="flex items-center gap-0.5 text-white/80 hover:text-white text-xs font-medium transition-colors"
        >
          전체 보기
          <ChevronRight size={14} />
        </Link>
      </div>

      {/* 일정 목록 (최대 3개) */}
      <div className="p-3 space-y-2">
        {remaining.slice(0, 3).map((task) => {
          const Icon = TYPE_ICON[task.schedule_type] ?? Star
          const colorCls = TYPE_COLOR[task.schedule_type] ?? TYPE_COLOR.custom
          const isCompleting = completingId === task.id

          return (
            <div
              key={task.id}
              className="flex items-center gap-3 px-1 py-1"
            >
              {/* 아이콘 */}
              <div
                className={clsx(
                  'w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0',
                  colorCls,
                )}
              >
                <Icon size={14} />
              </div>

              {/* 제목 */}
              <p className="flex-1 text-sm font-medium text-gray-800 truncate">{task.title}</p>

              {/* 빠른 완료 버튼 */}
              <button
                onClick={() => complete(task.id)}
                disabled={isCompleting}
                className={clsx(
                  'flex-shrink-0 w-7 h-7 rounded-full border-2 border-emerald-400',
                  'flex items-center justify-center hover:bg-emerald-50 active:scale-95',
                  'transition-all duration-150',
                  isCompleting && 'opacity-50 cursor-not-allowed',
                )}
                aria-label={`${task.title} 완료`}
              >
                {isCompleting ? (
                  <div className="w-3.5 h-3.5 border-2 border-emerald-400 border-t-transparent rounded-full animate-spin" />
                ) : (
                  <CheckCircle2 size={14} className="text-emerald-400" />
                )}
              </button>
            </div>
          )
        })}

        {/* 더 있으면 안내 */}
        {remaining.length > 3 && (
          <Link
            to="/care"
            className="block text-center text-xs text-gray-400 hover:text-primary-500 transition-colors pt-1"
          >
            + {remaining.length - 3}개 더 보기
          </Link>
        )}
      </div>
    </div>
  )
}
