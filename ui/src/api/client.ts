import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { useUIStore } from '@/stores/ui'

const ACCESS_KEY = 'azx_access'
const REFRESH_KEY = 'azx_refresh'

export const api: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

// One-time cleanup of poisoned localStorage from previous iterations
// (where the literal string "undefined" was accidentally stored as a
// token, then re-served on every request). The looksLikeJwt guard
// in the auth store and here means these values are silently dropped.
function looksLikeJwt(s: string | null | undefined): boolean {
  if (!s) return false
  const parts = s.split('.')
  return parts.length === 3 && parts.every(p => p.length > 0)
}

if (typeof window !== 'undefined') {
  try {
    for (const k of [ACCESS_KEY, REFRESH_KEY]) {
      const v = localStorage.getItem(k)
      if (v && !looksLikeJwt(v)) localStorage.removeItem(k)
    }
  } catch {}
}

function getAccessToken(): string | null {
  const v = localStorage.getItem(ACCESS_KEY)
  return looksLikeJwt(v) ? v : null
}

function attachAuthHeader(config: any) {
  const url = (config.url || '') as string
  // Don't attach to auth endpoints — they must be unauthenticated
  if (
    url.includes('/auth/login') ||
    url.includes('/auth/refresh') ||
    url.includes('/auth/setup')
  ) {
    if (config.headers && typeof config.headers.delete === 'function') {
      config.headers.delete('Authorization')
    } else if (config.headers) {
      delete config.headers.Authorization
    }
    return
  }
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
    const ui = useUIStore()
    const original = error.config as (InternalAxiosRequestConfig & { _retry?: boolean; _retried?: boolean }) | undefined

    const url = original?.url || ''
    const isAuthEndpoint =
      url.includes('/auth/login') ||
      url.includes('/auth/refresh') ||
      url.includes('/auth/setup')

    // For any 401 from a non-auth endpoint: blow away the local session
    // and force a re-login. No refresh attempts (they can loop, the
    // server is the source of truth, and 7-day access tokens are
    // long enough that silent re-login is fine).
    if (error.response?.status === 401 && original && !isAuthEndpoint) {
      if (typeof window !== 'undefined') {
        try {
          localStorage.removeItem(ACCESS_KEY)
          localStorage.removeItem(REFRESH_KEY)
          localStorage.removeItem('azx_user')
        } catch {}
        if (
          window.location.pathname !== '/login' &&
          window.location.pathname !== '/setup'
        ) {
          ui.showToast('Session expired — please sign in again', 'error')
          window.location.replace('/login')
        }
      }
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
