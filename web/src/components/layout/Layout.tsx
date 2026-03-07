import { useTranslation } from 'react-i18next'
import { Link, useLocation } from 'react-router-dom'
import { Fish, Users, ShoppingBag, Droplets, User, Globe } from 'lucide-react'
import { LOCALE_LABELS, SUPPORTED_LOCALES, isRTL } from '../../i18n'
import { clsx } from 'clsx'

interface LayoutProps {
  children: React.ReactNode
}

export default function Layout({ children }: LayoutProps) {
  const { t, i18n } = useTranslation()
  const location = useLocation()
  const rtl = isRTL(i18n.language)

  const navItems = [
    { to: '/', label: t('nav.home'), icon: null },
    { to: '/fish', label: t('nav.encyclopedia'), icon: Fish },
    { to: '/community', label: t('nav.community'), icon: Users },
    { to: '/marketplace', label: t('nav.marketplace'), icon: ShoppingBag },
    { to: '/tanks', label: t('nav.my_tanks'), icon: Droplets },
  ]

  return (
    <div className={clsx('min-h-screen bg-gray-50', rtl && 'font-rtl')} dir={rtl ? 'rtl' : 'ltr'}>
      {/* 상단 네비게이션 */}
      <header className="bg-white shadow-sm sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            {/* 로고 */}
            <Link to="/" className="flex items-center gap-2">
              <span className="text-2xl">🐠</span>
              <span className="text-xl font-bold text-primary-600">
                {t('app_name')}
              </span>
            </Link>

            {/* 데스크탑 네비 */}
            <nav className="hidden md:flex items-center gap-1">
              {navItems.map(({ to, label }) => (
                <Link
                  key={to}
                  to={to}
                  className={clsx(
                    'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                    location.pathname === to || (to !== '/' && location.pathname.startsWith(to))
                      ? 'bg-primary-50 text-primary-600'
                      : 'text-gray-600 hover:bg-gray-100',
                  )}
                >
                  {label}
                </Link>
              ))}
            </nav>

            {/* 우측: 로케일 + 로그인 */}
            <div className="flex items-center gap-3">
              {/* 언어 선택 */}
              <select
                value={i18n.language}
                onChange={(e) => i18n.changeLanguage(e.target.value)}
                className="text-sm border border-gray-200 rounded-lg px-2 py-1 bg-white"
              >
                {SUPPORTED_LOCALES.map((locale) => (
                  <option key={locale} value={locale}>
                    {LOCALE_LABELS[locale]}
                  </option>
                ))}
              </select>

              <Link
                to="/login"
                className="flex items-center gap-1.5 text-sm text-gray-600 hover:text-primary-600"
              >
                <User size={18} />
                {t('auth.login')}
              </Link>
            </div>
          </div>
        </div>
      </header>

      {/* 메인 콘텐츠 */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {children}
      </main>

      {/* 모바일 하단 탭바 */}
      <nav className="fixed bottom-0 left-0 right-0 bg-white border-t md:hidden z-50">
        <div className="flex">
          {navItems.map(({ to, label, icon: Icon }) => (
            <Link
              key={to}
              to={to}
              className={clsx(
                'flex-1 flex flex-col items-center py-2 text-xs',
                location.pathname === to || (to !== '/' && location.pathname.startsWith(to))
                  ? 'text-primary-600'
                  : 'text-gray-500',
              )}
            >
              {Icon && <Icon size={20} />}
              <span className="mt-0.5">{label}</span>
            </Link>
          ))}
        </div>
      </nav>
    </div>
  )
}
