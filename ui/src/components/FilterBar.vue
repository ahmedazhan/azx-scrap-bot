<script setup lang="ts">
import { Search, SlidersHorizontal, X } from 'lucide-vue-next'
import TypeBadge from './TypeBadge.vue'

export interface FilterValue {
  q: string
  sources: string[]
  types: string[]
  flagged_only: boolean
  from: string
  to: string
}

const props = defineProps<{
  modelValue: FilterValue
  sources: { key: string; name: string }[]
  types: { slug: string; label: string; dhivehi: string }[]
}>()
const emit = defineEmits<{
  (e: 'update:modelValue', v: FilterValue): void
  (e: 'open-mobile'): void
}>()

function update<K extends keyof FilterValue>(k: K, v: FilterValue[K]) {
  emit('update:modelValue', { ...props.modelValue, [k]: v })
}

function toggle(arr: string[], v: string) {
  return arr.includes(v) ? arr.filter((x) => x !== v) : [...arr, v]
}

const activeCount = () => {
  const f = props.modelValue
  return f.sources.length + f.types.length + (f.flagged_only ? 1 : 0) + (f.from ? 1 : 0) + (f.to ? 1 : 0)
}
</script>

<template>
  <div class="card flex flex-col gap-3 p-3 lg:flex-row lg:items-center lg:gap-2 lg:p-2">
    <div class="flex items-center gap-2 lg:flex-1">
      <div class="relative flex-1">
        <Search
          :size="14"
          :stroke-width="1.5"
          class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-dim"
        />
        <input
          :value="modelValue.q"
          @input="update('q', ($event.target as HTMLInputElement).value)"
          type="search"
          placeholder="Search items…"
          class="input-base pl-8"
        />
      </div>
      <button
        type="button"
        class="btn-ghost lg:hidden"
        @click="emit('open-mobile')"
      >
        <SlidersHorizontal :size="14" :stroke-width="1.5" />
        <span>Filters</span>
        <span v-if="activeCount() > 0" class="ml-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-aurora px-1 text-[10px] font-semibold text-white">
          {{ activeCount() }}
        </span>
      </button>
    </div>

    <div class="hidden flex-wrap items-center gap-2 lg:flex">
      <div class="flex items-center gap-1.5">
        <span class="text-[10px] uppercase tracking-wider text-dim">Source</span>
        <div class="flex flex-wrap gap-1">
          <button
            v-for="s in sources"
            :key="s.key"
            type="button"
            class="chip transition-colors duration-150 ease-out-expo"
            :class="modelValue.sources.includes(s.key) ? 'border-violet-400/40 bg-violet-400/10 text-thaana-text' : ''"
            @click="update('sources', toggle(modelValue.sources, s.key))"
          >
            {{ s.name }}
          </button>
        </div>
      </div>
      <div class="h-5 w-px bg-ink-700" />
      <div class="flex items-center gap-1.5">
        <span class="text-[10px] uppercase tracking-wider text-dim">Type</span>
        <div class="flex flex-wrap gap-1">
          <button
            v-for="t in types"
            :key="t.slug"
            type="button"
            class="transition-transform duration-150 ease-out-expo hover:scale-[1.02]"
            @click="update('types', toggle(modelValue.types, t.slug))"
          >
            <TypeBadge
              :slug="t.slug"
              :label="t.label"
              :dhivehi="t.dhivehi"
              :class="modelValue.types.includes(t.slug) ? 'ring-1 ring-aurora' : 'opacity-50'"
            />
          </button>
        </div>
      </div>
      <div class="h-5 w-px bg-ink-700" />
      <label class="flex cursor-pointer items-center gap-2 text-xs text-mid">
        <input
          type="checkbox"
          :checked="modelValue.flagged_only"
          @change="update('flagged_only', ($event.target as HTMLInputElement).checked)"
          class="h-3.5 w-3.5 rounded-sm border-ink-700 bg-transparent accent-violet-400"
        />
        Flagged only
      </label>
      <div class="h-5 w-px bg-ink-700" />
      <div class="flex items-center gap-1.5">
        <input
          :value="modelValue.from"
          @input="update('from', ($event.target as HTMLInputElement).value)"
          type="date"
          class="input-base h-8 w-36 px-2 text-xs"
        />
        <span class="text-dim">→</span>
        <input
          :value="modelValue.to"
          @input="update('to', ($event.target as HTMLInputElement).value)"
          type="date"
          class="input-base h-8 w-36 px-2 text-xs"
        />
      </div>
      <button
        v-if="activeCount() > 0"
        type="button"
        class="btn-ghost h-8 px-2 text-xs"
        @click="emit('update:modelValue', { q: '', sources: [], types: [], flagged_only: false, from: '', to: '' })"
      >
        <X :size="12" :stroke-width="1.75" />
        Clear
      </button>
    </div>
  </div>
</template>
