import { api } from './client'
import type { Dashboard } from './types'

export const dashboardApi = {
  async get() {
    const { data } = await api.get<Dashboard>('/dashboard')
    return data
  },
}
