<script setup lang="ts">
import { computed, ref } from 'vue'
import { useCountUp } from '@/composables/useCountUp'
import { TrendingUp, TrendingDown } from 'lucide-vue-next'

const props = defineProps<{
  label: string
  value: number
  hint?: string
  trend?: number
  accent?: 'mint' | 'violet' | 'pink' | 'sky' | 'teal' | 'amber'
}>()

const target = computed(() => props.value)
const display = useCountUp(target)

const accentClass = computed(() => {
  switch (props.accent) {
    case 'violet':
      return 'from-violet-400/30 to-violet-400/0'
    case 'mint':
      return 'from-mint-400/30 to-mint-400/0'
    case 'pink':
      return 'from-pink-400/30 to-pink-400/0'
    case 'sky':
      return 'from-sky-400/30 to-sky-400/0'
    case 'teal':
      return 'from-teal-400/30 to-teal-400/0'
    case 'amber':
      return 'from-signal-amber/30 to-signal-amber/0'
    default:
      return 'from-violet-400/30 to-violet-400/0'
  }
})

const trendPositive = computed(() => (props.trend ?? 0) >= 0)
</script>

<template>
  <div class="card relative overflow-hidden p-4">
    <div
      :class="['pointer-events-none absolute -right-8 -top-8 h-32 w-32 rounded-full bg-gradient-to-br blur-2xl', accentClass]"
    />
    <div class="relative flex flex-col gap-2">
      <div class="flex items-center justify-between">
        <span class="text-xs font-medium uppercase tracking-wider text-dim">{{ label }}</span>
        <div
          v-if="trend !== undefined"
          class="flex items-center gap-1 text-xs"
          :class="trendPositive ? 'text-signal-green' : 'text-signal-rose'"
        >
          <component :is="trendPositive ? TrendingUp : TrendingDown" :size="12" :stroke-width="1.75" />
          <span class="num">{{ Math.abs(trend) }}%</span>
        </div>
      </div>
      <div class="display num text-2xl font-semibold text-thaana-text sm:text-3xl">
        {{ display.toLocaleString() }}
      </div>
      <div v-if="hint" class="text-xs text-mid">{{ hint }}</div>
    </div>
  </div>
</template>
