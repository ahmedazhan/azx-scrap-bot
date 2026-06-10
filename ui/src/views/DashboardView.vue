<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import relativeTime from 'dayjs/plugin/relativeTime'
import {
  Activity,
  Flag,
  Radio,
  Inbox,
  CircleDot,
} from 'lucide-vue-next'
import { dashboardApi } from '@/api/dashboard'
import { sourcesApi } from '@/api/sources'
import { useUIStore } from '@/stores/ui'
import { useRouter } from 'vue-router'
import StatCard from '@/components/StatCard.vue'
import Sparkline from '@/components/Sparkline.vue'
import Skeleton from '@/components/Skeleton.vue'
import EmptyState from '@/components/EmptyState.vue'
import RadarMark from '@/components/RadarMark.vue'
import type { Dashboard, Source, LogEntry, FlagHit } from '@/api/types'

dayjs.extend(utc)
dayjs.extend(relativeTime)

const ui = useUIStore()
const router = useRouter()

const data = ref<Dashboard | null>(null)
const sources = ref<Source[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const recentFlags = ref<FlagHit[]>([])
const liveItems = ref<LogEntry[]>([])
const flashKeys = ref<Set<string>>(new Set())

function sum(arr: number[]) {
  return arr.reduce((a, b) => a + b, 0)
}

function pushLive(entry: LogEntry) {
  const key = `${entry.time}|${entry.msg}|${Math.random()}`
  flashKeys.value.add(key)
  flashKeys.value = new Set(flashKeys.value)
  liveItems.value = [entry, ...liveItems.value].slice(0, 50)
  setTimeout(() => {
    flashKeys.value.delete(key)
    flashKeys.value = new Set(flashKeys.value)
  }, 700)
}

let lastSeen = ''
watch(
  () => ui.recentLogs,
  (logs) => {
    if (!logs.length) return
    const newest = logs[0]
    const key = `${newest.time}|${newest.msg}`
    if (key === lastSeen) return
    lastSeen = key
    pushLive(newest)
  },
)

onMounted(async () => {
  try {
    const [d, s] = await Promise.all([dashboardApi.get(), sourcesApi.list()])
    data.value = d
    sources.value = s
    liveItems.value = d.recent_activity.slice(0, 50)
    const running = s.filter((x) => x.last_status === 'running').map((x) => x.key)
    ui.setActiveSources(running)
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
})

const totals = computed(() => data.value?.totals)

const sourcesHealth = computed(() => {
  const sl = data.value?.sparklines || {}
  return sources.value.map((s) => ({
    source: s,
    series: sl[s.key] || { fetched: new Array(24).fill(0), flagged: new Array(24).fill(0) },
  }))
})

function onKey(e: KeyboardEvent) {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
  if (e.key === 'g') {
    const next = (ev: KeyboardEvent) => {
      window.removeEventListener('keydown', next)
      if (ev.key === 'd') {
        const el = document.getElementById('kpi-strip')
        el?.focus()
        el?.scrollIntoView({ behavior: 'smooth', block: 'start' })
      }
    }
    window.addEventListener('keydown', next, { once: true })
    setTimeout(() => window.removeEventListener('keydown', next), 1500)
  }
}
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))

function relTime(s?: string | null) {
  if (!s) return '—'
  return dayjs.utc(s).fromNow()
}

function goFlagged() {
  router.push({ path: '/items', query: { flagged: '1' } })
}
</script>

<template>
  <div class="bg-noise min-h-full bg-ink-950 p-4 lg:p-6">
    <div class="mx-auto flex max-w-7xl flex-col gap-6">
      <header class="flex items-end justify-between">
        <div>
          <h1 class="display text-2xl font-semibold text-thaana-text">Dashboard</h1>
          <p class="mt-1 text-sm text-mid">Live snapshot of your scraping fleet.</p>
        </div>
        <div class="hidden items-center gap-1.5 rounded-md border border-ink-700 bg-ink-800/60 px-2 py-1 text-[10px] text-mid lg:flex">
          <span>Press</span>
          <kbd class="rounded bg-ink-700 px-1.5 py-0.5 font-mono text-[10px] text-thaana-text">g</kbd>
          <span>then</span>
          <kbd class="rounded bg-ink-700 px-1.5 py-0.5 font-mono text-[10px] text-thaana-text">d</kbd>
          <span>to focus KPIs</span>
        </div>
      </header>

      <div
        id="kpi-strip"
        tabindex="-1"
        class="grid grid-cols-1 gap-3 outline-none sm:grid-cols-2 lg:grid-cols-4"
      >
        <template v-if="loading">
          <Skeleton v-for="i in 4" :key="i" height="110px" rounded="0.5rem" />
        </template>
        <template v-else-if="totals">
          <StatCard label="Total items" :value="totals.items" :hint="`Across ${sources.length} source${sources.length === 1 ? '' : 's'}`" accent="violet" :icon="Inbox" />
          <StatCard label="New today" :value="totals.today" hint="Items fetched in last 24h" accent="mint" :trend="4" :icon="Activity" />
          <button type="button" class="text-left" @click="goFlagged">
            <StatCard label="Flagged unread" :value="totals.flagged" hint="Items matching your rules" accent="pink" :icon="Flag" />
          </button>
          <StatCard label="Sources on" :value="totals.sources_on" :hint="`${totals.sources_off} paused`" accent="teal" :icon="Radio" />
        </template>
      </div>

      <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <section class="card flex flex-col lg:col-span-2">
          <header class="flex items-center justify-between border-b border-ink-700 px-4 py-3">
            <h2 class="display text-sm font-semibold text-thaana-text">Live activity</h2>
            <div class="flex items-center gap-1.5 text-[10px] text-mid">
              <CircleDot :size="10" :stroke-width="2" class="text-signal-green" />
              streaming
            </div>
          </header>
          <div class="flex-1 overflow-hidden">
            <div v-if="loading" class="flex flex-col">
              <div v-for="i in 6" :key="i" class="flex items-center gap-3 border-b border-line px-4 py-3">
                <Skeleton width="44px" height="10px" />
                <Skeleton width="100%" height="12px" />
              </div>
            </div>
            <div v-else-if="!liveItems.length" class="p-4">
              <EmptyState variant="items" title="No activity yet" subtitle="Sources will appear here as they run." />
            </div>
            <ul v-else class="max-h-[420px] overflow-y-auto scroll-thin">
              <li
                v-for="(entry, idx) in liveItems"
                :key="`${entry.time}-${idx}-${entry.msg}`"
                class="flex items-start gap-3 border-b border-line px-4 py-2.5 text-sm"
              >
                <span
                  class="mt-0.5 inline-flex h-5 shrink-0 items-center rounded-md px-1.5 text-[10px] font-medium uppercase tracking-wider"
                  :class="{
                    'bg-signal-rose/15 text-signal-rose': entry.level === 'error',
                    'bg-signal-amber/15 text-signal-amber': entry.level === 'warn',
                    'bg-ink-700 text-mid': !['error', 'warn'].includes(entry.level),
                  }"
                >
                  {{ entry.level }}
                </span>
                <span class="min-w-0 flex-1 truncate text-thaana-text">{{ entry.msg }}</span>
                <span class="num shrink-0 text-[10px] text-dim">{{ relTime(entry.time) }}</span>
              </li>
            </ul>
          </div>
        </section>

        <section class="card flex flex-col">
          <header class="flex items-center justify-between border-b border-ink-700 px-4 py-3">
            <h2 class="display text-sm font-semibold text-thaana-text">Flagged queue</h2>
            <router-link to="/items?flagged=1" class="text-[10px] text-violet-400 hover:underline">View all</router-link>
          </header>
          <div class="flex-1">
            <div v-if="loading" class="flex flex-col">
              <div v-for="i in 4" :key="i" class="flex items-center gap-3 border-b border-line px-4 py-3">
                <Skeleton width="40%" height="12px" />
              </div>
            </div>
            <EmptyState
              v-else-if="!recentFlags.length"
              variant="flags"
              title="No flags"
              subtitle="Matching items will appear here."
            />
            <ul v-else class="max-h-[420px] overflow-y-auto scroll-thin">
              <li
                v-for="f in recentFlags"
                :key="f.id"
                class="flex items-center gap-3 border-b border-line px-4 py-2.5 text-sm transition-colors hover:bg-ink-800/50"
              >
                <Flag :size="12" :stroke-width="2" class="text-signal-rose" />
                <span class="min-w-0 flex-1 truncate">{{ f.item?.title || `Item #${f.item_id}` }}</span>
                <router-link :to="`/items/${f.item_id}`" class="shrink-0 text-[10px] text-violet-400 hover:underline">View</router-link>
              </li>
            </ul>
          </div>
        </section>
      </div>

      <section class="card">
        <header class="flex items-center justify-between border-b border-ink-700 px-4 py-3">
          <h2 class="display text-sm font-semibold text-thaana-text">Sources health</h2>
          <router-link to="/sources" class="text-[10px] text-violet-400 hover:underline">Manage</router-link>
        </header>
        <div v-if="loading" class="grid grid-cols-1 gap-3 p-4 sm:grid-cols-2 lg:grid-cols-3">
          <Skeleton v-for="i in 6" :key="i" height="120px" />
        </div>
        <div v-else class="grid grid-cols-1 gap-3 p-4 sm:grid-cols-2 lg:grid-cols-3">
          <div
            v-for="row in sourcesHealth"
            :key="row.source.key"
            class="flex flex-col gap-2 rounded-lg border border-ink-700 bg-ink-800/40 p-3"
          >
            <div class="flex items-center justify-between gap-2">
              <router-link
                :to="`/items?source=${row.source.key}`"
                class="display truncate text-[13px] font-medium text-thaana-text hover:text-violet-400"
              >
                {{ row.source.name }}
              </router-link>
              <div class="flex items-center gap-1.5">
                <RadarMark :active="row.source.last_status === 'running'" :size="12" />
                <span
                  class="inline-flex h-5 items-center rounded-md px-1.5 text-[10px] font-medium"
                  :class="{
                    'bg-signal-green/15 text-signal-green': row.source.last_status === 'ok' || !row.source.last_status,
                    'bg-signal-rose/15 text-signal-rose': row.source.last_status === 'error',
                    'bg-signal-amber/15 text-signal-amber': row.source.last_status === 'running',
                  }"
                >
                  {{ row.source.last_status || 'idle' }}
                </span>
              </div>
            </div>
            <div class="text-[10px] text-dim">Last run {{ relTime(row.source.last_run_at) }}</div>
            <div class="-mx-1">
              <Sparkline :fetched="row.series.fetched" :flagged="row.series.flagged" :height="32" :width="280" />
            </div>
            <div class="flex items-center justify-between text-[10px] text-dim">
              <span class="num">fetched <span class="text-mint-400">{{ sum(row.series.fetched) }}</span></span>
              <span class="num">flagged <span class="text-violet-400">{{ sum(row.series.flagged) }}</span></span>
              <span class="num">last 24h</span>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
