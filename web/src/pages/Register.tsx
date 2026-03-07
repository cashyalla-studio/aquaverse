import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useTranslation } from 'react-i18next'
import { Fish, Eye, EyeOff, UserPlus } from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import { LOCALE_LABELS, SUPPORTED_LOCALES } from '../i18n'

const registerSchema = z
  .object({
    email: z.string().email('Invalid email address'),
    password: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string(),
    nickname: z
      .string()
      .min(2, 'Nickname must be at least 2 characters')
      .max(20, 'Nickname must be at most 20 characters'),
    locale: z.string().min(1, 'Please select a language'),
    terms: z.literal(true, {
      errorMap: () => ({ message: 'You must accept the terms and conditions' }),
    }),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  })

type RegisterFormData = z.infer<typeof registerSchema>

export default function Register() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const registerFn = useAuthStore((s) => s.register)

  const [showPassword, setShowPassword] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [serverError, setServerError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      locale: (i18n.language as (typeof SUPPORTED_LOCALES)[number]) || 'en-US',
    },
  })

  const onSubmit = async (data: RegisterFormData) => {
    setServerError(null)
    try {
      await registerFn({
        email: data.email,
        password: data.password,
        nickname: data.nickname,
        locale: data.locale,
      })
      navigate('/login', { state: { registered: true } })
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } }
      setServerError(
        error?.response?.data?.message ||
          t('auth.registerError', 'Registration failed. Please try again.'),
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
            <p className="text-gray-500 text-sm mt-1">{t('auth.register')}</p>
          </div>

          {/* 에러 메시지 */}
          {serverError && (
            <div className="mb-5 px-4 py-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">
              {serverError}
            </div>
          )}

          {/* 폼 */}
          <form onSubmit={handleSubmit(onSubmit)} noValidate className="space-y-4">
            {/* 이메일 */}
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1.5">
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

            {/* 닉네임 */}
            <div>
              <label
                htmlFor="nickname"
                className="block text-sm font-medium text-gray-700 mb-1.5"
              >
                {t('auth.nickname')}
              </label>
              <input
                id="nickname"
                type="text"
                autoComplete="username"
                {...register('nickname')}
                className={`w-full px-4 py-2.5 border rounded-lg text-sm outline-none transition-colors
                  focus:ring-2 focus:ring-primary-200 focus:border-primary-400
                  ${errors.nickname ? 'border-red-400 bg-red-50' : 'border-gray-200 bg-white'}`}
                placeholder="AquaLover42"
              />
              {errors.nickname && (
                <p className="mt-1 text-xs text-red-500">{errors.nickname.message}</p>
              )}
            </div>

            {/* 비밀번호 */}
            <div>
              <label
                htmlFor="password"
                className="block text-sm font-medium text-gray-700 mb-1.5"
              >
                {t('auth.password')}
              </label>
              <div className="relative">
                <input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="new-password"
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

            {/* 비밀번호 확인 */}
            <div>
              <label
                htmlFor="confirmPassword"
                className="block text-sm font-medium text-gray-700 mb-1.5"
              >
                {t('auth.confirmPassword')}
              </label>
              <div className="relative">
                <input
                  id="confirmPassword"
                  type={showConfirm ? 'text' : 'password'}
                  autoComplete="new-password"
                  {...register('confirmPassword')}
                  className={`w-full px-4 py-2.5 pr-10 border rounded-lg text-sm outline-none transition-colors
                    focus:ring-2 focus:ring-primary-200 focus:border-primary-400
                    ${errors.confirmPassword ? 'border-red-400 bg-red-50' : 'border-gray-200 bg-white'}`}
                  placeholder="••••••••"
                />
                <button
                  type="button"
                  onClick={() => setShowConfirm((v) => !v)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                  aria-label={showConfirm ? 'Hide password' : 'Show password'}
                >
                  {showConfirm ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
              {errors.confirmPassword && (
                <p className="mt-1 text-xs text-red-500">{errors.confirmPassword.message}</p>
              )}
            </div>

            {/* 로케일 선택 */}
            <div>
              <label
                htmlFor="locale"
                className="block text-sm font-medium text-gray-700 mb-1.5"
              >
                {t('auth.locale')}
              </label>
              <select
                id="locale"
                {...register('locale')}
                className={`w-full px-4 py-2.5 border rounded-lg text-sm outline-none transition-colors bg-white
                  focus:ring-2 focus:ring-primary-200 focus:border-primary-400
                  ${errors.locale ? 'border-red-400 bg-red-50' : 'border-gray-200'}`}
              >
                {SUPPORTED_LOCALES.map((loc) => (
                  <option key={loc} value={loc}>
                    {LOCALE_LABELS[loc]}
                  </option>
                ))}
              </select>
              {errors.locale && (
                <p className="mt-1 text-xs text-red-500">{errors.locale.message}</p>
              )}
            </div>

            {/* 약관 동의 */}
            <div className="flex items-start gap-3 pt-1">
              <input
                id="terms"
                type="checkbox"
                {...register('terms')}
                className="mt-0.5 w-4 h-4 accent-primary-500 cursor-pointer"
              />
              <label htmlFor="terms" className="text-sm text-gray-600 cursor-pointer">
                {t('auth.termsAgree', 'I agree to the')}{' '}
                <a href="#" className="text-primary-500 hover:text-primary-700 font-medium">
                  {t('auth.termsOfService', 'Terms of Service')}
                </a>{' '}
                {t('auth.and', 'and')}{' '}
                <a href="#" className="text-primary-500 hover:text-primary-700 font-medium">
                  {t('auth.privacyPolicy', 'Privacy Policy')}
                </a>
              </label>
            </div>
            {errors.terms && (
              <p className="text-xs text-red-500 -mt-2">{errors.terms.message}</p>
            )}

            {/* 회원가입 버튼 */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-primary-500 hover:bg-primary-600
                disabled:bg-primary-300 text-white font-medium text-sm rounded-lg transition-colors mt-2"
            >
              {isSubmitting ? (
                <span className="w-4 h-4 border-2 border-white/40 border-t-white rounded-full animate-spin" />
              ) : (
                <UserPlus size={16} />
              )}
              {isSubmitting ? t('common.loading') : t('auth.register')}
            </button>
          </form>

          {/* 로그인 링크 */}
          <p className="mt-6 text-center text-sm text-gray-500">
            {t('auth.hasAccount')}{' '}
            <Link
              to="/login"
              className="text-primary-500 font-medium hover:text-primary-700 transition-colors"
            >
              {t('auth.login')}
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
