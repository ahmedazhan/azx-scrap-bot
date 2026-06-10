import { api } from './client'
import type { TelegramConfig, TelegramSubscriber } from './types'

export const telegramApi = {
  async getConfig() {
    const { data } = await api.get<{ config: TelegramConfig }>('/telegram/config')
    return data.config
  },
  async updateConfig(patch: Partial<TelegramConfig> & { bot_token?: string }) {
    const { data } = await api.put<{ config: TelegramConfig }>('/telegram/config', patch)
    return data.config
  },
  async testConnection() {
    const { data } = await api.post<{ ok: boolean; username?: string; error?: string }>(
      '/telegram/connect',
      {},
    )
    return data
  },
  async testSend() {
    const { data } = await api.post<{ results: { chat_id: string; ok: boolean; error?: string }[] }>(
      '/telegram/test',
      {},
    )
    return data.results
  },
  async listSubscribers() {
    const { data } = await api.get<TelegramSubscriber[]>('/telegram/subscribers')
    return data
  },
  async addSubscriber(chat_id: string, label: string) {
    const { data } = await api.post<{ subscriber: TelegramSubscriber }>(
      '/telegram/subscribers',
      { chat_id, label },
    )
    return data.subscriber
  },
  async updateSubscriber(id: number, patch: Partial<TelegramSubscriber>) {
    const { data } = await api.put<{ subscriber: TelegramSubscriber }>(
      `/telegram/subscribers/${id}`,
      patch,
    )
    return data.subscriber
  },
  async removeSubscriber(id: number) {
    const { data } = await api.delete<{ ok: boolean }>(`/telegram/subscribers/${id}`)
    return data
  },
}
