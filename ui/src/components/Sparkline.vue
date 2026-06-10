<script setup lang="ts">
import { computed, ref } from 'vue'

const props = withDefaults(
  defineProps<{
    fetched: number[]
    flagged: number[]
    height?: number
    width?: number
  }>(),
  { height: 36, width: 120 },
)

const tooltip = ref<{ x: number; y: number; f: number; fl: number; visible: boolean } | null>(null)

const padding = 2

const data = computed(() => {
  const f = props.fetched || []
  const fl = props.flagged || []
  const max = Math.max(1, ...f, ...fl)
  const w = props.width
  const h = props.height
  const stepX = f.length > 1 ? (w - padding * 2) / (f.length - 1) : 0

  const points = (arr: number[]) =>
    arr.map((v, i) => {
      const x = padding + i * stepX
      const y = h - padding - (v / max) * (h - padding * 2)
      return [x, y] as [number, number]
    })

  const toPath = (pts: [number, number][]) =>
    pts
      .map((p, i) => (i === 0 ? `M ${p[0]} ${p[1]}` : `L ${p[0]} ${p[1]}`))
      .join(' ')

  const toArea = (pts: [number, number][]) => {
    if (!pts.length) return ''
    const path = toPath(pts)
    return `${path} L ${pts[pts.length - 1][0]} ${h - padding} L ${pts[0][0]} ${h - padding} Z`
  }

  return {
    max,
    fetchedPath: toPath(points(f)),
    flaggedPath: toPath(points(fl)),
    fetchedArea: toArea(points(f)),
    flaggedArea: toArea(points(fl)),
    fetchedPoints: points(f),
    flaggedPoints: points(fl),
    w,
    h,
  }
})

function onMove(e: MouseEvent | TouchEvent) {
  const target = e.currentTarget as SVGElement
  const rect = target.getBoundingClientRect()
  const point = 'touches' in e ? e.touches[0] : e
  if (!point) return
  const x = point.clientX - rect.left
  const ratio = (x - padding) / (data.value.w - padding * 2)
  const idx = Math.max(0, Math.min(props.fetched.length - 1, Math.round(ratio * (props.fetched.length - 1))))
  const f = props.fetched[idx] ?? 0
  const fl = props.flagged[idx] ?? 0
  tooltip.value = {
    x: data.value.fetchedPoints[idx]?.[0] ?? 0,
    y: 0,
    f,
    fl,
    visible: true,
  }
}

function onLeave() {
  tooltip.value = null
}
</script>

<template>
  <div class="relative inline-block">
    <svg
      :width="data.w"
      :height="data.h"
      :viewBox="`0 0 ${data.w} ${data.h}`"
      class="block"
      @mousemove="onMove"
      @mouseleave="onLeave"
      @touchstart.passive="onMove"
      @touchend.passive="onLeave"
    >
      <defs>
        <linearGradient id="spark-fetched" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#3DDC97" stop-opacity="0.35" />
          <stop offset="100%" stop-color="#3DDC97" stop-opacity="0" />
        </linearGradient>
        <linearGradient id="spark-flagged" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#B968FF" stop-opacity="0.3" />
          <stop offset="100%" stop-color="#B968FF" stop-opacity="0" />
        </linearGradient>
      </defs>
      <path :d="data.fetchedArea" fill="url(#spark-fetched)" />
      <path :d="data.flaggedArea" fill="url(#spark-flagged)" />
      <path
        :d="data.fetchedPath"
        fill="none"
        stroke="#3DDC97"
        stroke-width="1.5"
        stroke-linejoin="round"
        stroke-linecap="round"
      />
      <path
        :d="data.flaggedPath"
        fill="none"
        stroke="#B968FF"
        stroke-width="1.5"
        stroke-dasharray="3 3"
        stroke-linejoin="round"
        stroke-linecap="round"
      />
    </svg>
    <div
      v-if="tooltip?.visible"
      class="pointer-events-none absolute -top-9 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-md border border-ink-700 bg-ink-900 px-2 py-1 text-[10px] num text-mid"
      :style="{ left: `${tooltip.x}px` }"
    >
      <span class="text-mint-400">{{ tooltip.f }}</span>
      <span class="mx-1 text-dim">·</span>
      <span class="text-violet-400">{{ tooltip.fl }}</span>
    </div>
  </div>
</template>
