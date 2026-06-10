import { api } from './client'
import type { Item, ItemFilters, Paginated } from './types'

export type ItemQuery = {
  source?: string[]
  type?: string[]
  q?: string
  flagged?: boolean
  from?: string
  to?: string
  page?: number
  page_size?: number
}

function qs(params: ItemQuery): string {
  const search = new URLSearchParams()
  if (params.source && params.source.length) search.set('source', params.source.join(','))
  if (params.type && params.type.length) search.set('type', params.type.join(','))
  if (params.q) search.set('q', params.q)
  if (params.flagged) search.set('flagged', '1')
  if (params.from) search.set('from', params.from)
  if (params.to) search.set('to', params.to)
  if (params.page) search.set('page', String(params.page))
  if (params.page_size) search.set('page_size', String(params.page_size))
  const s = search.toString()
  return s ? `?${s}` : ''
}

export const itemsApi = {
  async list(q: ItemQuery) {
    const { data } = await api.get<Paginated<Item>>(`/items${qs(q)}`)
    return data
  },
  async get(id: number | string) {
    const { data } = await api.get<{ item: Item }>(`/items/${id}`)
    return data.item
  },
  async filters() {
    const { data } = await api.get<ItemFilters>('/items/filters')
    return data
  },
  async markRead(id: number) {
    await api.post(`/items/${id}/read`, {})
  },
}
