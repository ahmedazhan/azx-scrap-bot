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
function looksLikeJwt(s: string | null | undefined): boolean {
  if (!s) return false
  const parts = s.split('.')
  return parts.length === 3 && parts.every(p => p.length > 0)
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

    // Skip if: not a 401, no original request, already retried once,
    // or this is an auth endpoint.
    if (
      error.response?.status !== 401 ||
      !original ||
      original._retried ||
      isAuthEndpoint
    ) {
      if (error.response && (error.response.status as number) >= 500) {
        ui.showToast('Server error — please retry', 'error')
      }
      return Promise.reject(error)
    }

    original._retried = true

    // Single-flight: one refresh for any number of concurrent 401s.
    if (!refreshing) {
      refreshing = (async () => {
        try {
          const stored = localStorage.getItem(REFRESH_KEY)
          if (!looksLikeJwt(stored)) {
            if (stored) localStorage.removeItem(REFRESH_KEY)
            return false
          }
          const res = await fetch((import.meta.env.VITE_API_BASE_URL || '/api') + '/auth/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh: stored }),
          })
          if (!res.ok) return false
          const json = await res.json()
          const data = json?.data
          if (!looksLikeJwt(data?.access) || !looksLikeJwt(data?.refresh)) return false
          localStorage.setItem(ACCESS_KEY, data.access)
          localStorage.setItem(REFRESH_KEY, data.refresh)
          return true
        } catch {
          return false
        }
      })()
    }
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

    // Refresh failed. Nuke localStorage and bounce to /login exactly once.
    const errBody = (error.response?.data as { error?: string } | undefined)?.error || ''
    if (typeof window !== 'undefined') {
      try {
        localStorage.removeItem(ACCESS_KEY)
        localStorage.removeItem(REFRESH_KEY)
        localStorage.removeItem('azx_user')
      } catch {}
      if (window.location.pathname !== '/login' && window.location.pathname !== '/setup') {
        ui.showToast('Session expired — please sign in again', 'error')
        window.location.replace('/login')
      }
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
