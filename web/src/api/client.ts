import axios, { AxiosInstance } from 'axios'
import i18n from '../i18n'

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

// 공유 axios 인스턴스
export const apiClient: AxiosInstance = axios.create({
  baseURL: `${BASE_URL}/api/v1`,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 요청 인터셉터: 인증 토큰 + 로케일 헤더 자동 추가
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('av_access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  // 현재 로케일을 X-Locale 헤더로 전달 (게시판 분리 핵심)
  config.headers['X-Locale'] = i18n.language || 'en-US'
  return config
})

// 응답 인터셉터: 401 시 토큰 갱신
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true
      try {
        const refreshToken = localStorage.getItem('av_refresh_token')
        const res = await axios.post(`${BASE_URL}/api/v1/auth/refresh`, {
          refresh_token: refreshToken,
        })
        const { access_token } = res.data
        localStorage.setItem('av_access_token', access_token)
        original.headers.Authorization = `Bearer ${access_token}`
        return apiClient(original)
      } catch {
        localStorage.removeItem('av_access_token')
        localStorage.removeItem('av_refresh_token')
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  },
)

export default apiClient
