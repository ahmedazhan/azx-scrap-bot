import { api } from './client'

export type SetupInfo = {
  setup_required: boolean
  user_count: number
  token_source: 'env' | 'db' | 'none'
  env_token_set: boolean
  env_admin_set: boolean
}

export const setupApi = {
  async info() {
    const { data } = await api.get<SetupInfo>('/setup-info')
    return data
  },
}
