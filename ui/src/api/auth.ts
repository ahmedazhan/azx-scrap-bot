import { api } from './client'
import type { User } from './types'

export const authApi = {
  async login(username: string, password: string) {
    const { data } = await api.post<{ access: string; refresh: string }>('/auth/login', {
      username,
      password,
    })
    return data
  },
  async refresh(refresh: string) {
    const { data } = await api.post<{ access: string; refresh: string }>('/auth/refresh', {
      refresh,
    })
    return data
  },
  async setup(username: string, password: string, setupToken: string) {
    const { data } = await api.post<{ access: string; refresh: string; user: User }>(
      '/auth/setup',
      {
        username,
        password,
        setup_token: setupToken,
      },
    )
    return data
  },
  async me() {
    const { data } = await api.get<{ user: User }>('/auth/me')
    return data
  },
  async changePassword(oldPassword: string, newPassword: string) {
    const { data } = await api.post<{ ok: boolean }>('/auth/change-password', {
      old: oldPassword,
      new: newPassword,
    })
    return data
  },
  async logout() {
    try {
      await api.post('/auth/logout', {})
    } catch {
      // ignore
    }
  },
}
