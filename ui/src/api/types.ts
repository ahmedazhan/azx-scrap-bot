export type User = {
  id?: number
  username: string
  theme?: 'dark' | 'light'
  filter_mode?: 'sheet' | 'inline'
  pull_to_refresh?: boolean
  created_at?: string
}

export type Source = {
  id?: number
  key: string
  name: string
  base_url: string
  enabled: boolean
  interval_sec: number
  concurrency: number
  max_pages_per_cycle: number
  last_run_at?: string | null
  last_status?: 'ok' | 'error' | 'running' | 'idle' | string
  last_error?: string | null
  created_at?: string
}

export type Item = {
  id: number
  source_key: string
  source_name?: string
  type_slug: string
  type_label?: string
  type_dhivehi?: string
  title: string
  title_dhivehi?: string
  url: string
  body?: string
  published_at?: string | null
  deadline_at?: string | null
  office?: string | null
  flagged: boolean
  unread: boolean
  flags?: FlagHit[]
  created_at?: string
}

export type FlagRule = {
  id: number
  name: string
  pattern: string
  is_regex: boolean
  match_in: 'title' | 'body' | 'url' | 'all'
  case_sensitive: boolean
  enabled: boolean
  last_match_count?: number
  last_match_at?: string | null
  created_at?: string
}

export type FlagHit = {
  id: number
  item_id: number
  rule_id: number
  rule_name?: string
  matched_at: string
  item?: Item
}

export type TelegramConfig = {
  enabled: boolean
  has_token: boolean
  bot_username?: string
  notify_on_flag_only: boolean
  throttle_ms: number
  updated_at?: string
}

export type TelegramSubscriber = {
  id: number
  chat_id: string
  label: string
  enabled: boolean
  added_at: string
  last_delivery_at?: string | null
  last_delivery_status?: 'ok' | 'error' | string | null
  last_delivery_error?: string | null
}

export type ScrapeRun = {
  id: number
  source_key: string
  started_at: string
  finished_at?: string | null
  status: 'running' | 'ok' | 'error'
  items_fetched: number
  items_new: number
  error?: string | null
}

export type DashboardTotals = {
  items: number
  today: number
  flagged: number
  sources_on: number
  sources_off: number
}

export type RecentActivity = {
  time: string
  level: 'info' | 'warn' | 'error' | 'debug' | string
  msg: string
  source?: string
  item_id?: number
  rule_id?: number
}

export type SparklineSeries = {
  fetched: number[]
  flagged: number[]
}

export type Dashboard = {
  totals: DashboardTotals
  recent_activity: RecentActivity[]
  sparklines: Record<string, SparklineSeries>
}

export type LogEntry = RecentActivity

export type Paginated<T> = {
  items: T[]
  total: number
  page: number
  page_size: number
}

export type ItemFilters = {
  sources: string[]
  types: { slug: string; label: string; dhivehi: string }[]
}
