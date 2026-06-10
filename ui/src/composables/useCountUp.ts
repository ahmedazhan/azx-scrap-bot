import { ref, watch, onUnmounted, type Ref } from 'vue'

export function useCountUp(target: Ref<number>, duration = 800) {
  const display = ref(0)
  let raf: number | null = null
  let start: number | null = null
  let from = 0
  let to = 0
  let active = false

  const reduceMotion =
    typeof window !== 'undefined' &&
    window.matchMedia &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches

  const ease = (t: number) => 1 - Math.pow(2, -10 * t)

  function step(ts: number) {
    if (start === null) start = ts
    const elapsed = ts - start
    const progress = Math.min(elapsed / duration, 1)
    display.value = Math.round(from + (to - from) * ease(progress))
    if (progress < 1) {
      raf = requestAnimationFrame(step)
    } else {
      raf = null
      start = null
      active = false
    }
  }

  watch(
    target,
    (v) => {
      if (reduceMotion) {
        display.value = v
        return
      }
      from = display.value
      to = v
      if (active && raf !== null) {
        start = null
        return
      }
      active = true
      raf = requestAnimationFrame(step)
    },
    { immediate: true },
  )

  onUnmounted(() => {
    if (raf !== null) cancelAnimationFrame(raf)
  })

  return display
}
