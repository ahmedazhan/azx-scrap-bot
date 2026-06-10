import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import relativeTime from 'dayjs/plugin/relativeTime'
import type { LogEntry } from '@/api/types'

dayjs.extend(utc)
dayjs.extend(relativeTime)

export type Theme = 'dark' | 'light'
export type FilterMode = 'sheet' | 'inline'

const THEME_KEY = 'azx_theme'
const FILTER_MODE_KEY = 'azx_filter_mode'
const DRAWER_KEY = 'azx_drawer'

export const useUIStore = defineStore('ui', () => {
  const theme = ref<Theme>((localStorage.getItem(THEME_KEY) as Theme) || 'dark')
  const filterMode = ref<FilterMode>(
    (localStorage.getItem(FILTER_MODE_KEY) as FilterMode) || 'sheet',
  )
  const drawerOpen = ref(false)
  const sidebarExpanded = ref(false)
  const recentLogs = ref<LogEntry[]>([])
  const activeSourceKeys = ref<Set<string>>(new Set())
  const toast = ref<{ msg: string; kind: 'info' | 'success' | 'error' } | null>(null)

  function setTheme(t: Theme) {
    theme.value = t
    localStorage.setItem(THEME_KEY, t)
    applyTheme()
  }

  function applyTheme() {
    if (theme.value === 'light') {
      document.documentElement.classList.remove('dark')
      document.documentElement.classList.add('light')
    } else {
      document.documentElement.classList.remove('light')
      document.documentElement.classList.add('dark')
    }
  }

  function setFilterMode(m: FilterMode) {
    filterMode.value = m
    localStorage.setItem(FILTER_MODE_KEY, m)
  }

  function setDrawer(open: boolean) {
    drawerOpen.value = open
    localStorage.setItem(DRAWER_KEY, open ? '1' : '0')
    if (open) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
  }

  function toggleDrawer() {
    setDrawer(!drawerOpen.value)
  }

  function setSidebarExpanded(v: boolean) {
    sidebarExpanded.value = v
  }

  function pushLog(entry: LogEntry) {
    recentLogs.value = [entry, ...recentLogs.value].slice(0, 200)
  }

  function seedLogs(entries: LogEntry[]) {
    recentLogs.value = entries.slice(0, 200)
  }

  function setActiveSources(keys: string[]) {
    activeSourceKeys.value = new Set(keys)
  }

  function showToast(msg: string, kind: 'info' | 'success' | 'error' = 'info') {
    toast.value = { msg, kind }
    setTimeout(() => {
      toast.value = null
    }, 2400)
  }

  const anySourceActive = computed(() => activeSourceKeys.value.size > 0)

  applyTheme()

  return {
    theme,
    filterMode,
    drawerOpen,
    sidebarExpanded,
    recentLogs,
    activeSourceKeys,
    toast,
    anySourceActive,
    setTheme,
    setFilterMode,
    setDrawer,
    toggleDrawer,
    setSidebarExpanded,
    pushLog,
    seedLogs,
    setActiveSources,
    showToast,
  }
})
