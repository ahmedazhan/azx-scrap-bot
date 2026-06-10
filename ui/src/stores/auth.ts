import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'
import { authApi } from '@/api/auth'
import type { User } from '@/api/types'

const ACCESS_KEY = 'azx_access'
const REFRESH_KEY = 'azx_refresh'
const USER_KEY = 'azx_user'

export const useAuthStore = defineStore('auth', () => {
  const access = ref<string | null>(localStorage.getItem(ACCESS_KEY))
  const refresh = ref<string | null>(localStorage.getItem(REFRESH_KEY))
  const user = ref<User | null>(
    localStorage.getItem(USER_KEY) ? JSON.parse(localStorage.getItem(USER_KEY)!) : null,
  )
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!access.value)

  function setTokens(a: string, r: string) {
    access.value = a
    refresh.value = r
    localStorage.setItem(ACCESS_KEY, a)
    localStorage.setItem(REFRESH_KEY, r)
  }

  function setUser(u: User | null) {
    user.value = u
    if (u) localStorage.setItem(USER_KEY, JSON.stringify(u))
    else localStorage.removeItem(USER_KEY)
  }

  async function login(username: string, password: string) {
    loading.value = true
    error.value = null
    try {
      const res = await authApi.login(username, password)
      setTokens(res.access, res.refresh)
      await fetchMe()
    } catch (e: unknown) {
      if (axios.isAxiosError(e)) {
        error.value = e.response?.data?.error || e.message
      } else {
        error.value = 'Login failed'
      }
      throw e
    } finally {
      loading.value = false
    }
  }

  async function setup(username: string, password: string, setupToken: string) {
    loading.value = true
    error.value = null
    try {
      const res = await authApi.setup(username, password, setupToken)
      setTokens(res.access, res.refresh)
      if (res.user) setUser(res.user)
    } catch (e: unknown) {
      if (axios.isAxiosError(e)) {
        error.value = e.response?.data?.error || e.message
      } else {
        error.value = 'Setup failed'
      }
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchMe() {
    try {
      const res = await authApi.me()
      setUser(res.user)
    } catch {
      setUser(null)
    }
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    await authApi.changePassword(oldPassword, newPassword)
  }

  function logout() {
    access.value = null
    refresh.value = null
    user.value = null
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
    localStorage.removeItem(USER_KEY)
  }

  async function tryRefresh(): Promise<boolean> {
    const stored = refresh.value || localStorage.getItem(REFRESH_KEY)
    if (!stored) {
      logout()
      return false
    }
    try {
      const res = await authApi.refresh(stored)
      setTokens(res.access, res.refresh)
      return true
    } catch {
      logout()
      return false
    }
  }

  return {
    access,
    refresh,
    user,
    loading,
    error,
    isAuthenticated,
    login,
    setup,
    fetchMe,
    changePassword,
    logout,
    tryRefresh,
    setUser,
  }
})
