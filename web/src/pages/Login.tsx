import { useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useTranslation } from 'react-i18next'
import { Fish, Eye, EyeOff, LogIn } from 'lucide-react'
import { useAuthStore } from '../store/authStore'

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
})

type LoginFormData = z.infer<typeof loginSchema>

export default function Login() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()
  const login = useAuthStore((s) => s.login)

  const [showPassword, setShowPassword] = useState(false)
  const [serverError, setServerError] = useState<string | null>(null)

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/'

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (data: LoginFormData) => {
    setServerError(null)
    try {
      await login(data.email, data.password)
      navigate(from, { replace: true })
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } }
      setServerError(
        error?.response?.data?.message || t('auth.loginError', 'Invalid email or password.'),
      )
    }
  }

  return (
    <div className="min-h-[calc(100vh-8rem)] flex items-center justify-center py-12 px-4">
      <div className="w-full max-w-md">
        {/* 카드 */}
        <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-8">
          {/* 로고 & 타이틀 */}
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-primary-50 rounded-2xl mb-4">
              <Fish className="text-primary-500" size={32} />
            </div>
            <h1 className="text-2xl font-bold text-gray-900">AquaVerse</h1>
            <p className="text-gray-500 text-sm mt-1">{t('auth.login')}</p>
          </div>

          {/* 에러 메시지 */}
          {serverError && (
            <div className="mb-5 px-4 py-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">
              {serverError}
            </div>
          )}

          {/* 폼 */}
          <form onSubmit={handleSubmit(onSubmit)} noValidate className="space-y-5">
            {/* 이메일 */}
            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium text-gray-700 mb-1.5"
              >
                {t('auth.email')}
              </label>
              <input
                id="email"
                type="email"
                autoComplete="email"
                {...register('email')}
                className={`w-full px-4 py-2.5 border rounded-lg text-sm outline-none transition-colors
                  focus:ring-2 focus:ring-primary-200 focus:border-primary-400
                  ${errors.email ? 'border-red-400 bg-red-50' : 'border-gray-200 bg-white'}`}
                placeholder="you@example.com"
              />
              {errors.email && (
                <p className="mt-1 text-xs text-red-500">{errors.email.message}</p>
              )}
            </div>

            {/* 비밀번호 */}
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                  {t('auth.password')}
                </label>
                <button
                  type="button"
                  className="text-xs text-primary-500 hover:text-primary-700 transition-colors"
                >
                  {t('auth.forgotPassword')}
                </button>
              </div>
              <div className="relative">
                <input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="current-password"
                  {...register('password')}
                  className={`w-full px-4 py-2.5 pr-10 border rounded-lg text-sm outline-none transition-colors
                    focus:ring-2 focus:ring-primary-200 focus:border-primary-400
                    ${errors.password ? 'border-red-400 bg-red-50' : 'border-gray-200 bg-white'}`}
                  placeholder="••••••••"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                >
                  {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
              {errors.password && (
                <p className="mt-1 text-xs text-red-500">{errors.password.message}</p>
              )}
            </div>

            {/* 로그인 버튼 */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-primary-500 hover:bg-primary-600
                disabled:bg-primary-300 text-white font-medium text-sm rounded-lg transition-colors"
            >
              {isSubmitting ? (
                <span className="w-4 h-4 border-2 border-white/40 border-t-white rounded-full animate-spin" />
              ) : (
                <LogIn size={16} />
              )}
              {isSubmitting ? t('common.loading') : t('auth.login')}
            </button>
          </form>

          {/* 회원가입 링크 */}
          <p className="mt-6 text-center text-sm text-gray-500">
            {t('auth.noAccount')}{' '}
            <Link
              to="/register"
              className="text-primary-500 font-medium hover:text-primary-700 transition-colors"
            >
              {t('auth.register')}
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
