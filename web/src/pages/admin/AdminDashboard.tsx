import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import api from '../../api/client'
import { useAuthStore } from '../../store/authStore'

interface KPI {
  mau: number
  dau: number
  total_users: number
  pro_subscribers: number
  total_listings: number
  active_trades: number
  escrow_success_rate: number
  escrow_disputes: number
  cites_filter_hits: number
  claude_api_calls_today: number
}

export default function AdminDashboard() {
  const [kpi, setKpi] = useState<KPI | null>(null)
  const { user } = useAuthStore()

  useEffect(() => {
    api.get('/admin/kpi').then(r => setKpi(r.data)).catch(() => {})
  }, [])

  if (user?.role !== 'ADMIN') {
    return <div className="p-8 text-red-500">접근 권한이 없습니다.</div>
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">관리자 대시보드</h1>

      {/* 네비게이션 */}
      <div className="flex gap-4 mb-8">
        <Link to="/admin/users" className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
          사용자 관리
        </Link>
        <Link to="/admin/audit" className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700">
          감사 로그
        </Link>
        <Link to="/admin/cites" className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700">
          CITES 통계
        </Link>
      </div>

      {/* KPI 그리드 */}
      {kpi ? (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <KPICard label="MAU" value={kpi.mau.toLocaleString()} color="blue" />
          <KPICard label="DAU" value={kpi.dau.toLocaleString()} color="indigo" />
          <KPICard label="전체 사용자" value={kpi.total_users.toLocaleString()} color="purple" />
          <KPICard label="PRO 구독자" value={kpi.pro_subscribers.toLocaleString()} color="yellow" />
          <KPICard label="전체 분양글" value={kpi.total_listings.toLocaleString()} color="green" />
          <KPICard label="진행 중 거래" value={kpi.active_trades.toLocaleString()} color="teal" />
          <KPICard
            label="에스크로 성공률"
            value={`${kpi.escrow_success_rate.toFixed(1)}%`}
            color={kpi.escrow_success_rate > 90 ? "green" : "orange"}
          />
          <KPICard
            label="에스크로 분쟁"
            value={kpi.escrow_disputes.toLocaleString()}
            color={kpi.escrow_disputes > 0 ? "red" : "gray"}
          />
          <KPICard label="CITES 차단 (오늘)" value={kpi.cites_filter_hits.toLocaleString()} color="orange" />
          <KPICard label="Claude API 호출 (오늘)" value={kpi.claude_api_calls_today.toLocaleString()} color="violet" />
        </div>
      ) : (
        <div className="text-gray-500">KPI 로딩 중...</div>
      )}
    </div>
  )
}

function KPICard({ label, value, color }: { label: string; value: string; color: string }) {
  const colorMap: Record<string, string> = {
    blue: 'bg-blue-50 border-blue-200 text-blue-700',
    indigo: 'bg-indigo-50 border-indigo-200 text-indigo-700',
    purple: 'bg-purple-50 border-purple-200 text-purple-700',
    yellow: 'bg-yellow-50 border-yellow-200 text-yellow-700',
    green: 'bg-green-50 border-green-200 text-green-700',
    teal: 'bg-teal-50 border-teal-200 text-teal-700',
    orange: 'bg-orange-50 border-orange-200 text-orange-700',
    red: 'bg-red-50 border-red-200 text-red-700',
    gray: 'bg-gray-50 border-gray-200 text-gray-700',
    violet: 'bg-violet-50 border-violet-200 text-violet-700',
  }
  return (
    <div className={`border rounded-lg p-4 ${colorMap[color] || colorMap.gray}`}>
      <div className="text-sm font-medium opacity-75">{label}</div>
      <div className="text-3xl font-bold mt-1">{value}</div>
    </div>
  )
}
