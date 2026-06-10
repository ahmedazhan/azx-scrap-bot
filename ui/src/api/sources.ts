import { api } from './client'
import type { Source } from './types'

export const sourcesApi = {
  async list() {
    const { data } = await api.get<Source[]>('/sources')
    return data
  },
  async update(
    key: string,
    patch: Partial<Pick<Source, 'enabled' | 'interval_sec' | 'concurrency' | 'max_pages_per_cycle'>>,
  ) {
    const { data } = await api.put<{ source: Source }>(`/sources/${key}`, patch)
    return data.source
  },
  async runNow(key: string) {
    const { data } = await api.post<{ ok: boolean }>(`/sources/${key}/run-now`, {})
    return data
  },
  async runAll() {
    const { data } = await api.post<{ ok: boolean }>('/sources/run-all', {})
    return data
  },
}
