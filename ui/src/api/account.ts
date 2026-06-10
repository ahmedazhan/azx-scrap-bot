import { api } from './client'
import type { User } from './types'

export const accountApi = {
  async get() {
    const { data } = await api.get<{ user: User }>('/account')
    return data.user
  },
  async update(patch: Partial<User>) {
    const { data } = await api.put<{ user: User }>('/account', patch)
    return data.user
  },
}
