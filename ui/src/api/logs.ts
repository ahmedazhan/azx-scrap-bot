import { api } from './client'
import type { LogEntry } from './types'

export const logsApi = {
  async recent(limit = 200) {
    const { data } = await api.get<LogEntry[]>(`/logs/recent?limit=${limit}`)
    return data
  },
}
