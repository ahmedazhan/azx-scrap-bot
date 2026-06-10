<script setup lang="ts">
import { computed } from 'vue'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import relativeTime from 'dayjs/plugin/relativeTime'
import TypeBadge from './TypeBadge.vue'
import { ExternalLink, Flag } from 'lucide-vue-next'
import type { Item } from '@/api/types'

dayjs.extend(utc)
dayjs.extend(relativeTime)

const props = defineProps<{ item: Item; compact?: boolean }>()
const emit = defineEmits<{ (e: 'click', item: Item): void }>()

const publishedRel = computed(() =>
  props.item.published_at ? dayjs.utc(props.item.published_at).fromNow() : '—',
)
const deadlineRel = computed(() => {
  if (!props.item.deadline_at) return null
  const d = dayjs.utc(props.item.deadline_at)
  const now = dayjs.utc()
  const days = d.diff(now, 'day')
  if (days < 0) return { text: `${Math.abs(days)}d overdue`, urgent: true }
  if (days <= 3) return { text: `${days}d left`, urgent: true }
  return { text: d.format('DD MMM'), urgent: false }
})
</script>

<template>
  <button
    type="button"
    class="group flex w-full flex-col gap-2 rounded-lg border border-ink-700 bg-ink-900/60 p-3 text-left transition-all duration-150 ease-out-expo hover:border-ink-600 hover:bg-ink-800/60"
    :class="compact ? 'p-2.5' : 'p-3'"
    @click="emit('click', item)"
  >
    <div class="flex items-start justify-between gap-2">
      <TypeBadge
        v-if="item.type_label"
        :slug="item.type_slug"
        :label="item.type_label"
        :dhivehi="item.type_dhivehi || ''"
      />
      <div class="flex items-center gap-1.5">
        <span
          v-if="item.flagged"
          class="inline-flex h-5 items-center gap-1 rounded-md bg-signal-rose/15 px-1.5 text-[10px] font-medium text-signal-rose"
        >
          <Flag :size="10" :stroke-width="2" />
          Flag
        </span>
        <span
          v-if="item.unread"
          class="h-1.5 w-1.5 rounded-full bg-signal-green"
          aria-label="unread"
        />
      </div>
    </div>
    <div class="display text-sm font-medium leading-snug text-thaana-text line-clamp-2">
      {{ item.title }}
    </div>
    <div v-if="item.title_dhivehi" dir="rtl" class="thaana line-clamp-1 text-xs text-mid">
      {{ item.title_dhivehi }}
    </div>
    <div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-mid">
      <span class="truncate">{{ item.source_name || item.source_key }}</span>
      <span class="num">{{ publishedRel }}</span>
      <span v-if="deadlineRel" class="num" :class="deadlineRel.urgent ? 'text-signal-amber' : ''">
        {{ deadlineRel.text }}
      </span>
      <ExternalLink
        v-if="item.url"
        :size="11"
        :stroke-width="1.5"
        class="ml-auto opacity-40 group-hover:opacity-80"
      />
    </div>
  </button>
</template>
