import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'

const ACCESS_KEY = 'azx_access'
const REFRESH_KEY = 'azx_refresh'

let refreshing: Promise<boolean> | null = null

export const api: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

// Read the access token directly from localStorage in the interceptor so
// the value is always current — the Pinia store's ref can transiently
// be empty during HMR or right after setTokens(), which would otherwise
// cause authed requests to land without an Authorization header.
function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_KEY)
}

function attachAuthHeader(config: any) {
  const token = getAccessToken()
  if (!token) return
  if (config.headers && typeof config.headers.set === 'function') {
    config.headers.set('Authorization', `Bearer ${token}`)
  } else {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
}

api.interceptors.request.use((config) => {
  attachAuthHeader(config)
  return config
})

api.interceptors.response.use(
  (r) => r,
  async (error: AxiosError) => {
    const auth = useAuthStore()
    const ui = useUIStore()
    const original = error.config as (InternalAxiosRequestConfig & { _retry?: boolean; _retried?: boolean }) | undefined

    const isAuthEndpoint =
      !!original?.url &&
      (original.url.includes('/auth/login') ||
        original.url.includes('/auth/refresh') ||
        original.url.includes('/auth/setup'))

    if (error.response?.status === 401 && original && !original._retried && !isAuthEndpoint) {
      original._retried = true
      // Try refresh once
      if (auth.refresh || localStorage.getItem(REFRESH_KEY)) {
        if (!refreshing) refreshing = auth.tryRefresh()
        const ok = await refreshing
        refreshing = null
        if (ok) {
          const newToken = localStorage.getItem(ACCESS_KEY)
          if (newToken) {
            if (original.headers && typeof (original.headers as any).set === 'function') {
              ;(original.headers as any).set('Authorization', `Bearer ${newToken}`)
            } else {
              original.headers = (original.headers as any) || {}
              ;(original.headers as any).Authorization = `Bearer ${newToken}`
            }
            return api.request(original)
          }
        }
      }
      // Refresh failed or no refresh token. The token (or secret) is bad.
      // Nuke all localStorage and reload to /login.
      const errBody = (error.response?.data as { error?: string } | undefined)?.error || ''
      if (typeof window !== 'undefined') {
        try {
          localStorage.removeItem(ACCESS_KEY)
          localStorage.removeItem(REFRESH_KEY)
          localStorage.removeItem('azx_user')
        } catch {}
        if (window.location.pathname !== '/login' && window.location.pathname !== '/setup') {
          if (errBody === 'invalid token' || errBody === 'invalid refresh') {
            ui.showToast('Session expired — please sign in again', 'error')
          }
          window.location.replace('/login')
        }
      }
      return Promise.reject(error)
    }

    if (error.response && (error.response.status as number) >= 500) {
      ui.showToast('Server error — please retry', 'error')
    }
    return Promise.reject(error)
  },
)

export function getErrorMessage(e: unknown): string {
  if (axios.isAxiosError(e)) {
    const data = e.response?.data as { error?: string; message?: string } | undefined
    return data?.error || data?.message || e.message
  }
  if (e instanceof Error) return e.message
  return 'Unknown error'
}
