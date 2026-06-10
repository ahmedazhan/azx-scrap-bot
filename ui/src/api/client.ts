import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'

const ACCESS_KEY = 'azx_access'
const REFRESH_KEY = 'azx_refresh'

let refreshing: Promise<boolean> | null = null

export const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

function attachAuthHeader(config: any) {
  const auth = useAuthStore()
  const token = auth.access || localStorage.getItem(ACCESS_KEY)
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
    const original = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    if (error.response?.status === 401 && !original._retry && auth.refresh) {
      original._retry = true
      if (!refreshing) refreshing = auth.tryRefresh()
      const ok = await refreshing
      refreshing = null
      if (ok && original) {
        if (original.headers && typeof (original.headers as any).set === 'function') {
          ;(original.headers as any).set('Authorization', `Bearer ${auth.access}`)
        } else {
          original.headers = (original.headers as any) || {}
          ;(original.headers as any).Authorization = `Bearer ${auth.access}`
        }
        return api.request(original)
      }
      auth.logout()
      if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
        window.location.href = '/login'
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
