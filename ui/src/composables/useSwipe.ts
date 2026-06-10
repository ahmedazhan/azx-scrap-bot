import { onMounted, onUnmounted, type Ref } from 'vue'

export type SwipeOptions = {
  onSwipeLeft?: () => void
  onSwipeRight?: () => void
  threshold?: number
}

export function useSwipe(target: Ref<HTMLElement | null>, opts: SwipeOptions) {
  const threshold = opts.threshold ?? 40
  let startX = 0
  let startY = 0
  let active = false

  const onStart = (e: TouchEvent) => {
    if (e.touches.length !== 1) return
    startX = e.touches[0].clientX
    startY = e.touches[0].clientY
    active = true
  }
  const onEnd = (e: TouchEvent) => {
    if (!active) return
    active = false
    const t = e.changedTouches[0]
    if (!t) return
    const dx = t.clientX - startX
    const dy = t.clientY - startY
    if (Math.abs(dy) > Math.abs(dx)) return
    if (Math.abs(dx) < threshold) return
    if (dx < 0) opts.onSwipeLeft?.()
    else opts.onSwipeRight?.()
  }

  onMounted(() => {
    const el = target.value
    if (!el) return
    el.addEventListener('touchstart', onStart, { passive: true })
    el.addEventListener('touchend', onEnd)
  })

  onUnmounted(() => {
    const el = target.value
    if (!el) return
    el.removeEventListener('touchstart', onStart)
    el.removeEventListener('touchend', onEnd)
  })
}
