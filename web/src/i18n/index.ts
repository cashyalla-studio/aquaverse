import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import Backend from 'i18next-http-backend'

// 지원 로케일 (13개)
export const SUPPORTED_LOCALES = [
  'ko',
  'en-US',
  'en-GB',
  'en-AU',
  'ja',
  'zh-CN',
  'zh-TW',
  'de',
  'fr-FR',
  'fr-CA',
  'es',
  'pt',
  'ar',
  'he',
] as const

export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number]

// RTL 언어
export const RTL_LOCALES: SupportedLocale[] = ['ar', 'he']

export function isRTL(locale: string): boolean {
  return RTL_LOCALES.includes(locale as SupportedLocale)
}

// 언어 표시명
export const LOCALE_LABELS: Record<SupportedLocale, string> = {
  ko: '한국어',
  'en-US': 'English (US)',
  'en-GB': 'English (UK)',
  'en-AU': 'English (AU)',
  ja: '日本語',
  'zh-CN': '中文 (简体)',
  'zh-TW': '中文 (繁體)',
  de: 'Deutsch',
  'fr-FR': 'Français (FR)',
  'fr-CA': 'Français (CA)',
  es: 'Español',
  pt: 'Português',
  ar: 'العربية',
  he: 'עברית',
}

i18n
  .use(Backend)
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    supportedLngs: SUPPORTED_LOCALES,
    fallbackLng: 'en-US',
    defaultNS: 'common',
    ns: ['common', 'fish', 'community', 'marketplace', 'auth'],

    backend: {
      loadPath: '/locales/{{lng}}/{{ns}}.json',
    },

    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      caches: ['localStorage'],
      lookupLocalStorage: 'av_locale',
    },

    interpolation: {
      escapeValue: false,
    },

    react: {
      useSuspense: true,
    },
  })

// 로케일 변경 시 HTML dir 속성 업데이트 (RTL 지원)
i18n.on('languageChanged', (lng) => {
  const dir = isRTL(lng) ? 'rtl' : 'ltr'
  document.documentElement.dir = dir
  document.documentElement.lang = lng
})

export default i18n
