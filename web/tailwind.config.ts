import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#0ea5e9',  // 메인 아쿠아 블루
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
        },
        tropical: {
          orange: '#f97316',
          coral: '#fb923c',
          teal: '#14b8a6',
          green: '#22c55e',
        },
      },
      fontFamily: {
        // RTL 언어 지원 폰트 스택
        sans: [
          'Inter',
          'Noto Sans',
          'Noto Sans Arabic',
          'Noto Sans Hebrew',
          'Noto Sans SC',
          'Noto Sans TC',
          'Noto Sans JP',
          'Noto Sans KR',
          'system-ui',
          'sans-serif',
        ],
      },
    },
  },
  plugins: [],
} satisfies Config
