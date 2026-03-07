import { useEffect, useState } from 'react'
import api from '../../api/client'

interface AdminUser {
  id: string
  email: string
  username: string
  role: string
  trust_score: number
  is_banned: boolean
  phone_verified: boolean
  created_at: string
  listing_count: number
  trade_count: number
  fraud_reports: number
}

export default function AdminUsers() {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(false)

  const load = async (q = '') => {
    setLoading(true)
    try {
      const r = await api.get('/admin/users', { params: { limit: 50, offset: 0, q } })
      setUsers(r.data.users || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleBan = async (userId: string, isBanned: boolean) => {
    if (isBanned) {
      await api.post(`/admin/users/${userId}/unban`)
    } else {
      const reason = prompt('정지 사유를 입력하세요:')
      if (!reason) return
      await api.post(`/admin/users/${userId}/ban`, { reason })
    }
    load(query)
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-2xl font-bold mb-4">사용자 관리</h1>
      <div className="flex gap-2 mb-4">
        <input
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="이메일 또는 닉네임 검색"
          className="flex-1 border rounded px-3 py-2"
          onKeyDown={e => e.key === 'Enter' && load(query)}
        />
        <button onClick={() => load(query)} className="px-4 py-2 bg-blue-600 text-white rounded">
          검색
        </button>
      </div>

      {loading ? <div>로딩 중...</div> : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm border-collapse">
            <thead>
              <tr className="bg-gray-50">
                <th className="border px-3 py-2 text-left">이메일</th>
                <th className="border px-3 py-2 text-left">닉네임</th>
                <th className="border px-3 py-2 text-center">역할</th>
                <th className="border px-3 py-2 text-center">신뢰도</th>
                <th className="border px-3 py-2 text-center">분양글</th>
                <th className="border px-3 py-2 text-center">거래</th>
                <th className="border px-3 py-2 text-center">신고</th>
                <th className="border px-3 py-2 text-center">전화인증</th>
                <th className="border px-3 py-2 text-center">상태</th>
                <th className="border px-3 py-2 text-center">관리</th>
              </tr>
            </thead>
            <tbody>
              {users.map(u => (
                <tr key={u.id} className={u.is_banned ? 'bg-red-50' : ''}>
                  <td className="border px-3 py-2">{u.email}</td>
                  <td className="border px-3 py-2">{u.username}</td>
                  <td className="border px-3 py-2 text-center">
                    <span className={`px-2 py-0.5 rounded text-xs ${u.role === 'ADMIN' ? 'bg-purple-100 text-purple-700' : 'bg-gray-100'}`}>
                      {u.role}
                    </span>
                  </td>
                  <td className="border px-3 py-2 text-center">{u.trust_score}</td>
                  <td className="border px-3 py-2 text-center">{u.listing_count}</td>
                  <td className="border px-3 py-2 text-center">{u.trade_count}</td>
                  <td className="border px-3 py-2 text-center">
                    <span className={u.fraud_reports > 0 ? 'text-red-600 font-bold' : ''}>{u.fraud_reports}</span>
                  </td>
                  <td className="border px-3 py-2 text-center">{u.phone_verified ? '✅' : '❌'}</td>
                  <td className="border px-3 py-2 text-center">
                    <span className={`px-2 py-0.5 rounded text-xs ${u.is_banned ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'}`}>
                      {u.is_banned ? '정지' : '정상'}
                    </span>
                  </td>
                  <td className="border px-3 py-2 text-center">
                    <button
                      onClick={() => handleBan(u.id, u.is_banned)}
                      className={`px-2 py-1 text-xs rounded ${u.is_banned ? 'bg-green-600 text-white' : 'bg-red-600 text-white'}`}
                    >
                      {u.is_banned ? '정지해제' : '정지'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
