import { api } from './client'
import type { FlagRule } from './types'

export type RuleTestPayload = {
  pattern: string
  is_regex: boolean
  match_in: FlagRule['match_in']
  case_sensitive: boolean
}

export type RuleTestResult = {
  matches: number
  total: number
  sample: { id: number; title: string; snippet: string }[]
}

export const rulesApi = {
  async list() {
    const { data } = await api.get<FlagRule[]>('/rules')
    return data
  },
  async create(rule: Omit<FlagRule, 'id' | 'created_at' | 'last_match_count' | 'last_match_at'>) {
    const { data } = await api.post<{ rule: FlagRule }>('/rules', rule)
    return data.rule
  },
  async update(id: number, patch: Partial<FlagRule>) {
    const { data } = await api.put<{ rule: FlagRule }>(`/rules/${id}`, patch)
    return data.rule
  },
  async remove(id: number) {
    const { data } = await api.delete<{ ok: boolean }>(`/rules/${id}`)
    return data
  },
  async test(payload: RuleTestPayload) {
    const { data } = await api.post<RuleTestResult>('/rules/test', payload)
    return data
  },
}
