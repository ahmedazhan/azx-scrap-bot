<script setup lang="ts">
import { ref, onMounted } from 'vue'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import relativeTime from 'dayjs/plugin/relativeTime'
import { Play, ExternalLink, AlertCircle } from 'lucide-vue-next'
import { sourcesApi } from '@/api/sources'
import { dashboardApi } from '@/api/dashboard'
import { useUIStore } from '@/stores/ui'
import Sparkline from '@/components/Sparkline.vue'
import RadarMark from '@/components/RadarMark.vue'
import Skeleton from '@/components/Skeleton.vue'
import EmptyState from '@/components/EmptyState.vue'
import type { Source, SparklineSeries } from '@/api/types'

dayjs.extend(utc)
dayjs.extend(relativeTime)

const sources = ref<Source[]>([])
const series = ref<Record<string, SparklineSeries>>({})
const loading = ref(true)
const running = ref<Set<string>>(new Set())
const errors = ref<Record<string, string | null>>({})
const ui = useUIStore()

onMounted(async () => {
  try {
    const [s, d] = await Promise.all([sourcesApi.list(), dashboardApi.get()])
    sources.value = s
    series.value = d.sparklines
  } finally {
    loading.value = false
  }
})

const rel = (s?: string | null) => (s ? dayjs.utc(s).fromNow() : 'never')

async function toggle(s: Source) {
  try {
    const updated = await sourcesApi.update(s.key, { enabled: !s.enabled })
    Object.assign(s, updated)
  } catch {
    ui.showToast('Failed to update source', 'error')
  }
}

async function updateInterval(s: Source, v: number) {
  try {
    const updated = await sourcesApi.update(s.key, { interval_sec: v })
    Object.assign(s, updated)
  } catch {
    ui.showToast('Failed to update interval', 'error')
  }
}

async function updateConcurrency(s: Source, v: number) {
  try {
    const updated = await sourcesApi.update(s.key, { concurrency: v })
    Object.assign(s, updated)
  } catch {
    ui.showToast('Failed to update concurrency', 'error')
  }
}

async function updateMaxPages(s: Source, v: number) {
  try {
    const updated = await sourcesApi.update(s.key, { max_pages_per_cycle: v })
    Object.assign(s, updated)
  } catch {
    ui.showToast('Failed to update max pages', 'error')
  }
}

async function runNow(s: Source) {
  if (running.value.has(s.key)) return
  running.value.add(s.key)
  running.value = new Set(running.value)
  errors.value[s.key] = null
  try {
    await sourcesApi.runNow(s.key)
    ui.showToast(`Triggered ${s.name}`, 'success')
    setTimeout(async () => {
      const fresh = await sourcesApi.list()
      const found = fresh.find((x) => x.key === s.key)
      if (found) Object.assign(s, found)
    }, 800)
  } catch (e) {
    errors.value[s.key] = String(e)
    ui.showToast(`Run failed: ${s.name}`, 'error')
  } finally {
    setTimeout(() => {
      running.value.delete(s.key)
      running.value = new Set(running.value)
    }, 1200)
  }
}

function seriesFor(s: Source) {
  return series.value[s.key] || { fetched: new Array(24).fill(0), flagged: new Array(24).fill(0) }
}

function sum(arr: number[]) {
  return arr.reduce((a, b) => a + b, 0)
}
</script>

<template>
  <div class="bg-noise min-h-full bg-ink-950 p-4 lg:p-6">
    <div class="mx-auto flex max-w-7xl flex-col gap-4">
      <header>
        <h1 class="display text-2xl font-semibold text-thaana-text">Sources</h1>
        <p class="mt-1 text-sm text-mid">Configure where Azx scrapes from and how often.</p>
      </header>

      <div v-if="loading" class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <Skeleton v-for="i in 6" :key="i" height="280px" />
      </div>

      <EmptyState
        v-else-if="!sources.length"
        variant="sources"
        title="No sources yet"
        subtitle="Add a source in your server config to start scraping."
      />

      <div v-else class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <article
          v-for="s in sources"
          :key="s.key"
          class="card flex flex-col gap-3 p-4"
        >
          <header class="flex items-start justify-between gap-2">
            <div class="min-w-0">
              <h3 class="display truncate text-[15px] font-semibold text-thaana-text">{{ s.name }}</h3>
              <a
                :href="s.base_url"
                target="_blank"
                rel="noreferrer"
                class="mt-0.5 inline-flex items-center gap-1 truncate text-[11px] text-mid hover:text-violet-400"
              >
                <span class="truncate">{{ s.base_url }}</span>
                <ExternalLink :size="10" :stroke-width="1.5" />
              </a>
            </div>
            <button
              type="button"
              role="switch"
              :aria-checked="s.enabled"
              class="relative h-5 w-9 shrink-0 rounded-full transition-colors duration-150"
              :class="s.enabled ? 'bg-aurora' : 'bg-ink-700'"
              @click="toggle(s)"
            >
              <span
                class="absolute top-0.5 h-4 w-4 rounded-full bg-white transition-transform duration-150 ease-out-expo"
                :class="s.enabled ? 'translate-x-4' : 'translate-x-0.5'"
              />
            </button>
          </header>

          <div class="-mx-1">
            <Sparkline
              :fetched="seriesFor(s).fetched"
              :flagged="seriesFor(s).flagged"
              :height="36"
              :width="320"
            />
            <div class="mt-1 flex items-center justify-between text-[10px] text-dim">
              <span class="num">fetched <span class="text-mint-400">{{ sum(seriesFor(s).fetched) }}</span></span>
              <span class="num">flagged <span class="text-violet-400">{{ sum(seriesFor(s).flagged) }}</span></span>
              <span class="num">last 24h</span>
            </div>
          </div>

          <div class="flex flex-col gap-2.5 text-xs">
            <label class="flex flex-col gap-1">
              <div class="flex items-center justify-between text-[10px] uppercase tracking-wider text-dim">
                <span>Interval</span>
                <span class="num text-thaana-text">{{ Math.round(s.interval_sec / 60) }}m</span>
              </div>
              <input
                type="range"
                min="60"
                max="86400"
                step="60"
                :value="s.interval_sec"
                @change="updateInterval(s, Number(($event.target as HTMLInputElement).value))"
                class="accent-violet-400"
              />
              <div class="flex justify-between text-[9px] text-dim">
                <span>1m</span>
                <span>24h</span>
              </div>
            </label>

            <div class="grid grid-cols-2 gap-2">
              <label class="flex flex-col gap-1">
                <div class="text-[10px] uppercase tracking-wider text-dim">Concurrency</div>
                <input
                  type="number"
                  min="1"
                  max="32"
                  :value="s.concurrency"
                  @change="updateConcurrency(s, Number(($event.target as HTMLInputElement).value))"
                  class="input-base h-8 text-xs"
                />
              </label>
              <label class="flex flex-col gap-1">
                <div class="text-[10px] uppercase tracking-wider text-dim">Max pages</div>
                <input
                  type="number"
                  min="1"
                  max="500"
                  :value="s.max_pages_per_cycle"
                  @change="updateMaxPages(s, Number(($event.target as HTMLInputElement).value))"
                  class="input-base h-8 text-xs"
                />
              </label>
            </div>
          </div>

          <div v-if="errors[s.key]" class="flex items-start gap-2 rounded-md border border-signal-rose/30 bg-signal-rose/10 px-2 py-1.5 text-[11px] text-signal-rose">
            <AlertCircle :size="12" :stroke-width="2" class="mt-0.5 shrink-0" />
            <span class="line-clamp-2">{{ errors[s.key] }}</span>
          </div>

          <footer class="flex items-center justify-between gap-2 border-t border-line pt-3">
            <div class="flex items-center gap-1.5 text-[11px] text-mid">
              <RadarMark :active="running.has(s.key) || s.last_status === 'running'" :size="12" />
              <span>Last run {{ rel(s.last_run_at) }}</span>
            </div>
            <button
              class="btn-primary"
              :disabled="running.has(s.key)"
              @click="runNow(s)"
            >
              <Play v-if="!running.has(s.key)" :size="12" :stroke-width="2" />
              <RadarMark v-else :active="true" :size="12" />
              <span>{{ running.has(s.key) ? 'Running…' : 'Run now' }}</span>
            </button>
          </footer>
        </article>
      </div>
    </div>
  </div>
</template>
