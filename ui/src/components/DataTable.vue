<script setup lang="ts">
import { computed, ref } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

type Column<T> = {
  key: string
  label: string
  width?: string
  align?: 'left' | 'right' | 'center'
  render?: (row: T) => unknown
  class?: string
}

const props = defineProps<{
  columns: Column<unknown>[]
  rows: unknown[]
  rowKey?: string
  loading?: boolean
  emptyTitle?: string
  emptySubtitle?: string
}>()

const emit = defineEmits<{ (e: 'row-click', row: unknown): void }>()

function get(row: unknown, key: string): unknown {
  if (row && typeof row === 'object' && key in (row as Record<string, unknown>)) {
    return (row as Record<string, unknown>)[key]
  }
  return undefined
}

const gridStyle = computed(() => {
  return {
    gridTemplateColumns: props.columns
      .map((c) => c.width || (c.align === 'right' ? 'auto' : '1fr'))
      .join(' '),
  }
})
</script>

<template>
  <div class="card overflow-hidden">
    <div class="hidden lg:block">
      <div class="grid gap-3 border-b border-line bg-ink-900/60 px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-dim" :style="gridStyle">
        <div
          v-for="c in columns"
          :key="c.key"
          :class="[c.align === 'right' ? 'text-right' : c.align === 'center' ? 'text-center' : 'text-left', c.class]"
        >
          {{ c.label }}
        </div>
      </div>
      <div v-if="loading" class="divide-y divide-line">
        <div v-for="i in 6" :key="i" class="grid items-center gap-3 px-4 py-3.5" :style="gridStyle">
          <Skeleton v-for="c in columns" :key="c.key" :height="'14px'" :rounded="'0.25rem'" />
        </div>
      </div>
      <div v-else-if="!rows.length" class="px-6 py-12 text-center text-sm text-mid">
        {{ emptyTitle || 'No results' }}
      </div>
      <div v-else class="divide-y divide-line">
        <div
          v-for="(row, i) in rows"
          :key="String(get(row, rowKey || 'id') ?? i)"
          class="grid cursor-pointer items-center gap-3 px-4 py-3 text-sm transition-colors duration-150 ease-out-expo hover:bg-ink-800/50"
          :style="gridStyle"
          @click="emit('row-click', row)"
        >
          <div
            v-for="c in columns"
            :key="c.key"
            :class="[c.align === 'right' ? 'text-right' : c.align === 'center' ? 'text-center' : 'text-left', c.class, 'min-w-0 truncate']"
          >
            <slot :name="`cell-${c.key}`" :row="row" :value="get(row, c.key)">
              <component v-if="c.render" :is="c.render(row)" />
              <template v-else>{{ get(row, c.key) }}</template>
            </slot>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
