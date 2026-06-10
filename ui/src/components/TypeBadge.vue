<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  slug: string
  label: string
  dhivehi: string
}>()

const PALETTE = [
  { bg: 'rgba(185, 104, 255, 0.12)', fg: '#B968FF', border: 'rgba(185, 104, 255, 0.25)' },
  { bg: 'rgba(94, 196, 255, 0.12)', fg: '#5EC4FF', border: 'rgba(94, 196, 255, 0.25)' },
  { bg: 'rgba(79, 240, 200, 0.12)', fg: '#4FF0C8', border: 'rgba(79, 240, 200, 0.25)' },
  { bg: 'rgba(255, 107, 157, 0.12)', fg: '#FF6B9D', border: 'rgba(255, 107, 157, 0.25)' },
  { bg: 'rgba(255, 180, 84, 0.12)', fg: '#FFB454', border: 'rgba(255, 180, 84, 0.25)' },
  { bg: 'rgba(61, 220, 151, 0.12)', fg: '#3DDC97', border: 'rgba(61, 220, 151, 0.25)' },
]

const color = computed(() => {
  let h = 0
  for (let i = 0; i < props.slug.length; i++) h = (h * 31 + props.slug.charCodeAt(i)) | 0
  return PALETTE[Math.abs(h) % PALETTE.length]
})
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 text-[11px] font-medium"
    :style="{ backgroundColor: color.bg, color: color.fg, borderColor: color.border }"
  >
    <span class="leading-none">{{ label }}</span>
    <span v-if="dhivehi" dir="rtl" class="thaana text-[10px] opacity-70 leading-none">{{ dhivehi }}</span>
  </span>
</template>
