import { useTranslation } from 'react-i18next'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import {
  Fish,
  Users,
  ShoppingBag,
  Home,
  User,
  LogOut,
  Search,
  Bell,
  Video,
  Map,
  ChevronDown,
  HeartPulse,
} from 'lucide-react'
import ChatBot from '../ChatBot'
import { LOCALE_LABELS, SUPPORTED_LOCALES, isRTL } from '../../i18n'
import { clsx } from 'clsx'
import { useAuthStore } from '../../store/authStore'
import { useState } from 'react'

interface LayoutProps {
  children: React.ReactNode
}

const BOTTOM_TABS = [
  { to: '/', label: '홈', icon: Home, exact: true },
  { to: '/fish', label: '탐색', icon: Fish },
  { to: '/community', label: '커뮤니티', icon: Users },
  { to: '/marketplace', label: '마켓', icon: ShoppingBag },
  { to: '/care', label: '케어', icon: HeartPulse },
]

const TOP_NAV = [
  { to: '/', label: '홈', exact: true },
  { to: '/fish', label: '백과사전' },
  { to: '/community', label: '커뮤니티' },
  { to: '/marketplace', label: '마켓플레이스' },
  { to: '/care', label: '케어허브' },
  { to: '/videos', label: '영상' },
  { to: '/map', label: '지도' },
  { to: '/identify', label: 'AI 식별' },
  { to: '/badges', label: '뱃지' },
]

function isActive(to: string, pathname: string, exact = false) {
  if (exact) return pathname === to
  return pathname.startsWith(to) && to !== '/'
}

export default function Layout({ children }: LayoutProps) {
  const { i18n } = useTranslation()
  const location = useLocation()
  const navigate = useNavigate()
  const rtl = isRTL(i18n.language)
  const { isAuthenticated, user, logout } = useAuthStore()
  const [langOpen, setLangOpen] = useState(false)

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  return (
    <div
      className={clsx('min-h-screen bg-gray-50', rtl && 'font-rtl')}
      dir={rtl ? 'rtl' : 'ltr'}
    >
      {/* ──── 상단 헤더 ──── */}
      <header className="bg-white/95 backdrop-blur-sm border-b border-gray-100 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16 gap-4">
            {/* 로고 */}
            <Link to="/" className="flex items-center gap-2 flex-shrink-0">
              <span className="text-2xl leading-none">🐠</span>
              <span className="text-xl font-black tracking-tight text-gradient hidden sm:block">
                Finara
              </span>
            </Link>

            {/* 데스크탑 네비 */}
            <nav className="hidden lg:flex items-center gap-0.5">
              {TOP_NAV.map(({ to, label, exact }) => (
                <Link
                  key={to}
                  to={to}
                  className={clsx(
                    'px-3.5 py-2 rounded-xl text-sm font-medium transition-colors',
                    isActive(to, location.pathname, exact)
                      ? 'bg-primary-50 text-primary-600'
                      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900',
                  )}
                >
                  {label}
                </Link>
              ))}
            </nav>

            {/* 우측 액션 */}
            <div className="flex items-center gap-2">
              {/* 검색 버튼 (모바일) */}
              <Link
                to="/fish"
                className="lg:hidden p-2 rounded-xl text-gray-500 hover:bg-gray-100 transition-colors"
                aria-label="검색"
              >
                <Search size={20} />
              </Link>

              {/* AI 식별 버튼 */}
              <Link
                to="/identify"
                className="hidden sm:flex items-center gap-1.5 bg-gradient-to-r from-primary-500 to-cyan-500
                           text-white text-xs font-semibold px-3.5 py-2 rounded-xl hover:opacity-90
                           transition-opacity shadow-sm"
              >
                ✨ AI 식별
              </Link>

              {/* 언어 선택 */}
              <div className="relative">
                <button
                  onClick={() => setLangOpen(!langOpen)}
                  className="flex items-center gap-1 text-sm text-gray-600 border border-gray-200
                             rounded-xl px-2.5 py-1.5 hover:bg-gray-50 transition-colors"
                >
                  <span className="text-xs">{i18n.language.split('-')[0].toUpperCase()}</span>
                  <ChevronDown size={12} />
                </button>
                {langOpen && (
                  <div className="absolute right-0 mt-1 w-44 bg-white border border-gray-100 rounded-xl
                                  shadow-lg py-1 z-50 max-h-64 overflow-y-auto">
                    {SUPPORTED_LOCALES.map((locale) => (
                      <button
                        key={locale}
                        onClick={() => { i18n.changeLanguage(locale); setLangOpen(false) }}
                        className={clsx(
                          'w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 transition-colors',
                          i18n.language === locale && 'text-primary-600 font-medium',
                        )}
                      >
                        {LOCALE_LABELS[locale]}
                      </button>
                    ))}
                  </div>
                )}
              </div>

              {/* 인증 */}
              {isAuthenticated && user ? (
                <div className="flex items-center gap-1.5">
                  <Link
                    to="/feed"
                    className="hidden sm:flex p-2 rounded-xl text-gray-500 hover:bg-gray-100 transition-colors"
                    aria-label="알림"
                  >
                    <Bell size={18} />
                  </Link>
                  <div className="flex items-center gap-2 pl-1">
                    <span className="hidden sm:block w-8 h-8 rounded-full bg-gradient-to-br from-primary-400
                                     to-cyan-400 flex items-center justify-center text-white text-xs font-bold">
                      {user.nickname?.[0]?.toUpperCase() ?? 'U'}
                    </span>
                    <button
                      onClick={handleLogout}
                      className="p-2 rounded-xl text-gray-400 hover:text-red-500 hover:bg-red-50 transition-colors"
                      aria-label="로그아웃"
                    >
                      <LogOut size={16} />
                    </button>
                  </div>
                </div>
              ) : (
                <Link
                  to="/login"
                  className="flex items-center gap-1.5 text-sm font-medium text-gray-700
                             border border-gray-200 rounded-xl px-3.5 py-1.5 hover:bg-gray-50 transition-colors"
                >
                  <User size={15} />
                  <span className="hidden sm:inline">로그인</span>
                </Link>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* ──── 메인 콘텐츠 ──── */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 pb-24 lg:pb-8">
        {children}
      </main>

      {/* ──── AI 챗봇 FAB ──── */}
      <ChatBot />

      {/* ──── 모바일 하단 탭바 ──── */}
      <nav className="fixed bottom-0 left-0 right-0 bg-white/95 backdrop-blur-sm border-t border-gray-100 lg:hidden z-50 safe-area-inset-bottom">
        <div className="flex">
          {BOTTOM_TABS.map(({ to, label, icon: Icon, exact }) => {
            const active = exact
              ? location.pathname === to
              : location.pathname.startsWith(to) && to !== '/'
            return (
              <Link
                key={to}
                to={to}
                className={clsx(
                  'flex-1 flex flex-col items-center py-2.5 gap-0.5 text-xs font-medium transition-colors',
                  active ? 'text-primary-600' : 'text-gray-400 hover:text-gray-600',
                )}
              >
                <span
                  className={clsx(
                    'p-1 rounded-xl transition-colors',
                    active && 'bg-primary-50',
                  )}
                >
                  <Icon size={20} strokeWidth={active ? 2.5 : 1.8} />
                </span>
                <span className="leading-none">{label}</span>
              </Link>
            )
          })}
        </div>
      </nav>

      {/* 랭귀지 드롭다운 오버레이 */}
      {langOpen && (
        <div
          className="fixed inset-0 z-40"
          onClick={() => setLangOpen(false)}
        />
      )}
    </div>
  )
}
