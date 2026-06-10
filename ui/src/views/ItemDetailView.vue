<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import relativeTime from 'dayjs/plugin/relativeTime'
import { ArrowLeft, ExternalLink, Flag, Calendar, Building2, Radio } from 'lucide-vue-next'
import { itemsApi } from '@/api/items'
import TypeBadge from '@/components/TypeBadge.vue'
import Skeleton from '@/components/Skeleton.vue'
import ThaanaText from '@/components/ThaanaText.vue'
import type { Item } from '@/api/types'

dayjs.extend(utc)
dayjs.extend(relativeTime)

const route = useRoute()
const router = useRouter()

const item = ref<Item | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  try {
    const id = Number(route.params.id)
    item.value = await itemsApi.get(id)
    try {
      await itemsApi.markRead(id)
    } catch {
      // ignore
    }
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
})

const fmt = (s?: string | null) => (s ? dayjs.utc(s).format('DD MMM YYYY · HH:mm [UTC]') : '—')
const rel = (s?: string | null) => (s ? dayjs.utc(s).fromNow() : '—')
</script>

<template>
  <div class="bg-noise min-h-full bg-ink-950 p-4 lg:p-6">
    <div class="mx-auto flex max-w-3xl flex-col gap-4">
      <button class="btn-ghost self-start" @click="router.back()">
        <ArrowLeft :size="14" :stroke-width="1.75" />
        Back
      </button>

      <template v-if="loading">
        <Skeleton height="40px" />
        <Skeleton height="20px" width="60%" />
        <Skeleton height="240px" />
      </template>

      <div v-else-if="error" class="card p-6 text-sm text-signal-rose">
        {{ error }}
      </div>

      <article v-else-if="item" class="card overflow-hidden">
        <header class="flex flex-col gap-3 border-b border-ink-700 p-5">
          <div class="flex flex-wrap items-center gap-2">
            <TypeBadge
              v-if="item.type_label"
              :slug="item.type_slug"
              :label="item.type_label"
              :dhivehi="item.type_dhivehi || ''"
            />
            <span class="chip">
              <Radio :size="10" :stroke-width="1.5" />
              {{ item.source_name || item.source_key }}
            </span>
            <span
              v-if="item.flagged"
              class="inline-flex h-5 items-center gap-1 rounded-md bg-signal-rose/15 px-1.5 text-[10px] font-medium text-signal-rose"
            >
              <Flag :size="10" :stroke-width="2" />
              Flagged
            </span>
          </div>
          <h1 class="display text-xl font-semibold leading-snug text-thaana-text lg:text-2xl">
            {{ item.title }}
          </h1>
          <ThaanaText v-if="item.title_dhivehi" as="p" :text="item.title_dhivehi" class="block text-sm text-mid" />
        </header>

        <div class="grid grid-cols-2 gap-4 border-b border-ink-700 p-5 text-xs sm:grid-cols-4">
          <div class="flex flex-col gap-1">
            <div class="flex items-center gap-1.5 text-[10px] uppercase tracking-wider text-dim">
              <Calendar :size="10" :stroke-width="1.5" />
              Published
            </div>
            <div class="num text-thaana-text">{{ fmt(item.published_at) }}</div>
            <div class="num text-[10px] text-mid">{{ rel(item.published_at) }}</div>
          </div>
          <div v-if="item.deadline_at" class="flex flex-col gap-1">
            <div class="flex items-center gap-1.5 text-[10px] uppercase tracking-wider text-dim">
              <Calendar :size="10" :stroke-width="1.5" />
              Deadline
            </div>
            <div class="num text-thaana-text">{{ fmt(item.deadline_at) }}</div>
            <div class="num text-[10px] text-mid">{{ rel(item.deadline_at) }}</div>
          </div>
          <div v-if="item.office" class="flex flex-col gap-1">
            <div class="flex items-center gap-1.5 text-[10px] uppercase tracking-wider text-dim">
              <Building2 :size="10" :stroke-width="1.5" />
              Office
            </div>
            <div class="text-thaana-text">{{ item.office }}</div>
          </div>
          <div class="flex flex-col gap-1">
            <div class="text-[10px] uppercase tracking-wider text-dim">Source</div>
            <div class="text-thaana-text">{{ item.source_name || item.source_key }}</div>
          </div>
        </div>

        <div v-if="item.body" class="p-5">
          <div class="whitespace-pre-wrap text-sm leading-relaxed text-thaana-text">{{ item.body }}</div>
        </div>

        <footer class="flex items-center gap-2 border-t border-ink-700 p-4">
          <a
            v-if="item.url"
            :href="item.url"
            target="_blank"
            rel="noreferrer"
            class="btn-primary"
          >
            <ExternalLink :size="14" :stroke-width="1.75" />
            Open original
          </a>
          <button class="btn-ghost" @click="router.push('/items')">Back to list</button>
        </footer>
      </article>
    </div>
  </div>
</template>
