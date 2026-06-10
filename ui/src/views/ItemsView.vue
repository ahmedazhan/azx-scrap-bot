<script setup lang="ts">
import { ref, computed, onMounted, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import relativeTime from 'dayjs/plugin/relativeTime'
import { Search, ChevronLeft, ChevronRight, RefreshCw, SlidersHorizontal, X } from 'lucide-vue-next'
import { itemsApi, type ItemQuery } from '@/api/items'
import { sourcesApi } from '@/api/sources'
import { useUIStore } from '@/stores/ui'
import { useAuthStore } from '@/stores/auth'
import ItemCard from '@/components/ItemCard.vue'
import ItemDrawer from '@/components/ItemDrawer.vue'
import FilterBar, { type FilterValue } from '@/components/FilterBar.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import DataTable from '@/components/DataTable.vue'
import TypeBadge from '@/components/TypeBadge.vue'
import EmptyState from '@/components/EmptyState.vue'
import Skeleton from '@/components/Skeleton.vue'
import StickyActionBar from '@/components/StickyActionBar.vue'
import type { Item, ItemFilters, Source } from '@/api/types'

dayjs.extend(utc)
dayjs.extend(relativeTime)

const route = useRoute()
const router = useRouter()
const ui = useUIStore()
const auth = useAuthStore()

const items = ref<Item[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(25)
const loading = ref(false)
const filtersData = ref<ItemFilters>({ sources: [], types: [] })
const sources = ref<Source[]>([])
const drawerItem = ref<Item | null>(null)
const sheetOpen = ref(false)
const searchEl = ref<HTMLInputElement | null>(null)
const pulling = ref(false)
const pullStart = ref(0)

const filter = ref<FilterValue>({
  q: '',
  sources: [],
  types: [],
  flagged_only: false,
  from: '',
  to: '',
})

function readQuery() {
  filter.value.q = (route.query.q as string) || ''
  filter.value.sources = ((route.query.source as string) || '').split(',').filter(Boolean)
  filter.value.types = ((route.query.type as string) || '').split(',').filter(Boolean)
  filter.value.flagged_only = route.query.flagged === '1'
  filter.value.from = (route.query.from as string) || ''
  filter.value.to = (route.query.to as string) || ''
  page.value = Number(route.query.page) || 1
}

function writeQuery() {
  const q: Record<string, string> = {}
  if (filter.value.q) q.q = filter.value.q
  if (filter.value.sources.length) q.source = filter.value.sources.join(',')
  if (filter.value.types.length) q.type = filter.value.types.join(',')
  if (filter.value.flagged_only) q.flagged = '1'
  if (filter.value.from) q.from = filter.value.from
  if (filter.value.to) q.to = filter.value.to
  if (page.value > 1) q.page = String(page.value)
  router.replace({ query: q })
}

async function load() {
  loading.value = true
  try {
    const params: ItemQuery = {
      q: filter.value.q || undefined,
      source: filter.value.sources,
      type: filter.value.types,
      flagged: filter.value.flagged_only || undefined,
      from: filter.value.from || undefined,
      to: filter.value.to || undefined,
      page: page.value,
      page_size: pageSize.value,
    }
    const res = await itemsApi.list(params)
    items.value = res.items
    total.value = res.total
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  readQuery()
  const [f, s] = await Promise.all([itemsApi.filters(), sourcesApi.list()])
  filtersData.value = f
  sources.value = s
  await load()
  window.addEventListener('keydown', onKey)
})

onUnmounted(() => window.removeEventListener('keydown', onKey))

watch(filter, () => {
  page.value = 1
  writeQuery()
  load()
}, { deep: true })

watch(page, () => {
  writeQuery()
  load()
})

watch(() => route.query, () => {
  const next = {
    q: (route.query.q as string) || '',
    sources: ((route.query.source as string) || '').split(',').filter(Boolean),
    types: ((route.query.type as string) || '').split(',').filter(Boolean),
    flagged_only: route.query.flagged === '1',
    from: (route.query.from as string) || '',
    to: (route.query.to as string) || '',
  }
  if (JSON.stringify(next) !== JSON.stringify(filter.value)) {
    filter.value = next
    page.value = Number(route.query.page) || 1
    load()
  }
})

function onKey(e: KeyboardEvent) {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
  if (e.key === '/') {
    e.preventDefault()
    searchEl.value?.focus()
  }
}

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

const columns = [
  { key: 'type', label: 'Type', width: '160px' },
  { key: 'title', label: 'Title', width: 'minmax(0, 1fr)' },
  { key: 'source', label: 'Source', width: '160px' },
  { key: 'published', label: 'Published', width: '120px', align: 'right' as const },
  { key: 'deadline', label: 'Deadline', width: '120px', align: 'right' as const },
  { key: 'office', label: 'Office', width: '160px' },
  { key: 'actions', label: '', width: '80px', align: 'right' as const },
]

function sourceName(item: Item) {
  return item.source_name || item.source_key
}

function relTime(s?: string | null) {
  if (!s) return '—'
  return dayjs.utc(s).fromNow()
}

function onItemClick(item: Item) {
  if (window.innerWidth >= 1024) {
    drawerItem.value = item
    router.replace({ query: { ...route.query, item: String(item.id) } })
  } else {
    router.push(`/items/${item.id}`)
  }
}

watch(drawerItem, (v) => {
  if (!v) {
    const q = { ...route.query }
    delete (q as Record<string, unknown>).item
    router.replace({ query: q })
  }
})

onMounted(() => {
  const idParam = route.query.item
  if (idParam && items.value.length) {
    const found = items.value.find((i) => String(i.id) === String(idParam))
    if (found) drawerItem.value = found
  }
})

watch(items, (list) => {
  if (!drawerItem.value && route.query.item) {
    const found = list.find((i) => String(i.id) === String(route.query.item))
    if (found) drawerItem.value = found
  }
})

function nextPage() {
  if (page.value < totalPages.value) page.value++
}
function prevPage() {
  if (page.value > 1) page.value--
}

function openFilters() {
  if (ui.filterMode === 'sheet' || window.innerWidth < 1024) sheetOpen.value = true
}

const pullEnabled = computed(() => auth.user?.pull_to_refresh ?? false)

function onTouchStart(e: TouchEvent) {
  if (!pullEnabled.value) return
  if (window.scrollY > 0) return
  pullStart.value = e.touches[0].clientY
}
async function onTouchEnd(e: TouchEvent) {
  if (!pullEnabled.value) return
  const end = e.changedTouches[0].clientY
  if (end - pullStart.value > 80 && !pulling.value) {
    pulling.value = true
    try {
      const { sourcesApi } = await import('@/api/sources')
      await sourcesApi.runAll()
      ui.showToast('Refreshing all sources…', 'info')
      await load()
    } catch {
      ui.showToast('Refresh failed', 'error')
    } finally {
      setTimeout(() => (pulling.value = false), 600)
    }
  }
}

const activeTypeLabel = computed(() => {
  if (!filter.value.types.length) return null
  return filter.value.types
    .map((s) => filtersData.value.types.find((t) => t.slug === s)?.label || s)
    .join(', ')
})
</script>

<template>
  <div
    class="bg-noise min-h-full bg-ink-950 p-4 lg:p-6"
    @touchstart.passive="onTouchStart"
    @touchend.passive="onTouchEnd"
  >
    <div class="mx-auto flex max-w-7xl flex-col gap-4">
      <header class="flex flex-wrap items-end justify-between gap-2">
        <div>
          <h1 class="display text-2xl font-semibold text-thaana-text">Items</h1>
          <p class="mt-1 text-sm text-mid">
            <span class="num">{{ total.toLocaleString() }}</span> results
            <span v-if="activeTypeLabel"> · <span class="text-thaana-text">{{ activeTypeLabel }}</span></span>
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button class="btn-ghost lg:hidden" @click="openFilters">
            <SlidersHorizontal :size="14" :stroke-width="1.5" />
            <span>Filters</span>
          </button>
        </div>
      </header>

      <Transition
        enter-active-class="transition duration-150 ease-out-expo"
        leave-active-class="transition duration-150 ease-out-expo"
        enter-from-class="opacity-0 -translate-y-1"
        leave-to-class="opacity-0 -translate-y-1"
      >
        <div
          v-if="pulling"
          class="flex items-center justify-center gap-2 rounded-lg border border-ink-700 bg-ink-800/60 px-3 py-2 text-xs text-mid"
        >
          <RefreshCw :size="12" :stroke-width="2" class="animate-spin text-violet-400" />
          Refreshing all sources…
        </div>
      </Transition>

      <FilterBar v-model="filter" :sources="filtersData.sources.map((k) => ({ key: k, name: k }))" :types="filtersData.types" @open-mobile="sheetOpen = true" />

      <div class="hidden lg:block">
        <DataTable :columns="columns" :rows="items" :loading="loading" row-key="id" @row-click="(r) => onItemClick(r as Item)">
          <template #cell-type="{ row }">
            <TypeBadge
              v-if="(row as Item).type_label"
              :slug="(row as Item).type_slug"
              :label="(row as Item).type_label || ''"
              :dhivehi="(row as Item).type_dhivehi || ''"
            />
          </template>
          <template #cell-title="{ row }">
            <div class="flex flex-col">
              <span class="display truncate text-thaana-text">{{ (row as Item).title }}</span>
              <span v-if="(row as Item).title_dhivehi" dir="rtl" class="thaana truncate text-[11px] text-mid">
                {{ (row as Item).title_dhivehi }}
              </span>
            </div>
          </template>
          <template #cell-source="{ row }">
            <span class="truncate text-mid">{{ sourceName(row as Item) }}</span>
          </template>
          <template #cell-published="{ row }">
            <span class="num text-mid">{{ relTime((row as Item).published_at) }}</span>
          </template>
          <template #cell-deadline="{ row }">
            <span class="num text-mid">{{ relTime((row as Item).deadline_at) }}</span>
          </template>
          <template #cell-office="{ row }">
            <span class="truncate text-mid">{{ (row as Item).office || '—' }}</span>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end gap-1.5">
              <span
                v-if="(row as Item).flagged"
                class="inline-flex h-5 items-center gap-1 rounded-md bg-signal-rose/15 px-1.5 text-[10px] font-medium text-signal-rose"
              >
                flag
              </span>
              <span v-if="(row as Item).unread" class="h-1.5 w-1.5 rounded-full bg-signal-green" />
            </div>
          </template>
        </DataTable>
      </div>

      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:hidden">
        <template v-if="loading">
          <Skeleton v-for="i in 6" :key="i" height="120px" />
        </template>
        <template v-else-if="!items.length">
          <EmptyState variant="items" title="No items match" subtitle="Try adjusting your filters." />
        </template>
        <ItemCard
          v-for="item in items"
          :key="item.id"
          :item="item"
          @click="onItemClick(item)"
        />
      </div>

      <div class="flex items-center justify-between gap-2 border-t border-line pt-3">
        <div class="text-xs text-mid">
          Page <span class="num text-thaana-text">{{ page }}</span> of <span class="num text-thaana-text">{{ totalPages.toLocaleString() }}</span>
        </div>
        <div class="flex items-center gap-1.5">
          <button class="btn-ghost" :disabled="page <= 1" @click="prevPage">
            <ChevronLeft :size="14" :stroke-width="1.75" />
            <span class="hidden sm:inline">Prev</span>
          </button>
          <button class="btn-ghost" :disabled="page >= totalPages" @click="nextPage">
            <span class="hidden sm:inline">Next</span>
            <ChevronRight :size="14" :stroke-width="1.75" />
          </button>
        </div>
      </div>

      <StickyActionBar>
        <div class="flex items-center justify-between gap-2">
          <div class="text-xs text-mid">
            <span class="num">{{ total.toLocaleString() }}</span> results
          </div>
          <div class="flex items-center gap-1.5">
            <button class="btn-ghost" :disabled="page <= 1" @click="prevPage">
              <ChevronLeft :size="14" :stroke-width="1.75" />
            </button>
            <span class="num text-xs text-thaana-text">{{ page }} / {{ totalPages.toLocaleString() }}</span>
            <button class="btn-ghost" :disabled="page >= totalPages" @click="nextPage">
              <ChevronRight :size="14" :stroke-width="1.75" />
            </button>
          </div>
        </div>
      </StickyActionBar>
    </div>

    <ItemDrawer :open="!!drawerItem" :title="drawerItem?.title || ''" @close="drawerItem = null">
      <div v-if="drawerItem" class="flex flex-col gap-4 p-5">
        <div class="flex flex-wrap items-center gap-2">
          <TypeBadge
            v-if="drawerItem.type_label"
            :slug="drawerItem.type_slug"
            :label="drawerItem.type_label"
            :dhivehi="drawerItem.type_dhivehi || ''"
          />
          <span class="chip">{{ drawerItem.source_name || drawerItem.source_key }}</span>
          <span v-if="drawerItem.flagged" class="inline-flex h-5 items-center gap-1 rounded-md bg-signal-rose/15 px-1.5 text-[10px] font-medium text-signal-rose">
            Flagged
          </span>
        </div>
        <h1 class="display text-lg font-semibold leading-snug text-thaana-text">{{ drawerItem.title }}</h1>
        <p v-if="drawerItem.title_dhivehi" dir="rtl" class="thaana text-sm text-mid">
          {{ drawerItem.title_dhivehi }}
        </p>
        <div class="grid grid-cols-2 gap-3 text-xs">
          <div>
            <div class="text-[10px] uppercase tracking-wider text-dim">Published</div>
            <div class="num text-thaana-text">{{ relTime(drawerItem.published_at) }}</div>
          </div>
          <div v-if="drawerItem.deadline_at">
            <div class="text-[10px] uppercase tracking-wider text-dim">Deadline</div>
            <div class="num text-thaana-text">{{ relTime(drawerItem.deadline_at) }}</div>
          </div>
          <div v-if="drawerItem.office">
            <div class="text-[10px] uppercase tracking-wider text-dim">Office</div>
            <div class="text-thaana-text">{{ drawerItem.office }}</div>
          </div>
          <div>
            <div class="text-[10px] uppercase tracking-wider text-dim">Source</div>
            <div class="text-thaana-text">{{ drawerItem.source_name || drawerItem.source_key }}</div>
          </div>
        </div>
        <div v-if="drawerItem.body" class="prose prose-invert max-w-none whitespace-pre-wrap text-sm text-thaana-text">
          {{ drawerItem.body }}
        </div>
        <a
          v-if="drawerItem.url"
          :href="drawerItem.url"
          target="_blank"
          rel="noreferrer"
          class="btn-primary mt-2 self-start"
        >
          Open original
        </a>
      </div>
    </ItemDrawer>

    <BottomSheet :open="sheetOpen" title="Filters" @close="sheetOpen = false">
      <div class="flex flex-col gap-5 p-4">
        <div class="flex flex-col gap-2">
          <div class="text-[10px] font-medium uppercase tracking-wider text-dim">Source</div>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="k in filtersData.sources"
              :key="k"
              type="button"
              class="chip transition-colors"
              :class="filter.sources.includes(k) ? 'border-violet-400/40 bg-violet-400/10 text-thaana-text' : ''"
              @click="filter.sources = filter.sources.includes(k) ? filter.sources.filter((x) => x !== k) : [...filter.sources, k]"
            >
              {{ k }}
            </button>
          </div>
        </div>
        <div class="flex flex-col gap-2">
          <div class="text-[10px] font-medium uppercase tracking-wider text-dim">Type</div>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="t in filtersData.types"
              :key="t.slug"
              type="button"
              class="transition-transform hover:scale-[1.02]"
              @click="filter.types = filter.types.includes(t.slug) ? filter.types.filter((x) => x !== t.slug) : [...filter.types, t.slug]"
            >
              <TypeBadge
                :slug="t.slug"
                :label="t.label"
                :dhivehi="t.dhivehi"
                :class="filter.types.includes(t.slug) ? 'ring-1 ring-aurora' : 'opacity-50'"
              />
            </button>
          </div>
        </div>
        <label class="flex items-center gap-2 text-sm text-mid">
          <input
            type="checkbox"
            :checked="filter.flagged_only"
            @change="filter.flagged_only = ($event.target as HTMLInputElement).checked"
            class="h-4 w-4 rounded-sm border-ink-700 bg-transparent accent-violet-400"
          />
          Flagged only
        </label>
        <div class="grid grid-cols-2 gap-2">
          <div class="flex flex-col gap-1.5">
            <div class="text-[10px] font-medium uppercase tracking-wider text-dim">From</div>
            <input v-model="filter.from" type="date" class="input-base" />
          </div>
          <div class="flex flex-col gap-1.5">
            <div class="text-[10px] font-medium uppercase tracking-wider text-dim">To</div>
            <input v-model="filter.to" type="date" class="input-base" />
          </div>
        </div>
        <div class="flex items-center justify-between gap-2 border-t border-line pt-4">
          <button
            class="btn-ghost"
            @click="filter = { q: '', sources: [], types: [], flagged_only: false, from: '', to: '' }"
          >
            <X :size="14" :stroke-width="1.75" />
            Clear
          </button>
          <button class="btn-primary" @click="sheetOpen = false">Apply</button>
        </div>
      </div>
    </BottomSheet>
  </div>
</template>
