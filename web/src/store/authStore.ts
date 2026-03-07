import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { apiClient } from '../api/client'

export interface AuthUser {
  id: string
  nickname: string
  email: string
  role: string
}

export interface RegisterData {
  email: string
  password: string
  nickname: string
  locale: string
}

interface LoginResponse {
  access_token: string
  refresh_token: string
  user: AuthUser
}

interface AuthState {
  user: AuthUser | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  register: (data: RegisterData) => Promise<void>
  logout: () => void
  setTokens: (access: string, refresh: string) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,

      login: async (email: string, password: string) => {
        const res = await apiClient.post<LoginResponse>('/auth/login', {
          email,
          password,
        })
        const { access_token, refresh_token, user } = res.data
        localStorage.setItem('av_access_token', access_token)
        localStorage.setItem('av_refresh_token', refresh_token)
        set({
          user,
          accessToken: access_token,
          refreshToken: refresh_token,
          isAuthenticated: true,
        })
      },

      register: async (data: RegisterData) => {
        await apiClient.post('/auth/register', data)
      },

      logout: () => {
        localStorage.removeItem('av_access_token')
        localStorage.removeItem('av_refresh_token')
        set({
          user: null,
          accessToken: null,
          refreshToken: null,
          isAuthenticated: false,
        })
      },

      setTokens: (access: string, refresh: string) => {
        localStorage.setItem('av_access_token', access)
        localStorage.setItem('av_refresh_token', refresh)
        set({ accessToken: access, refreshToken: refresh })
      },
    }),
    {
      name: 'av_auth',
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    },
  ),
)
