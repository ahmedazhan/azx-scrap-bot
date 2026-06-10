import { ref, onMounted, onUnmounted, type Ref } from 'vue'

export function useMediaQuery(query: string): Ref<boolean> {
  const matches = ref(false)
  let mql: MediaQueryList | null = null
  const onChange = (e: MediaQueryListEvent) => {
    matches.value = e.matches
  }

  onMounted(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return
    mql = window.matchMedia(query)
    matches.value = mql.matches
    if (mql.addEventListener) mql.addEventListener('change', onChange)
    else mql.addListener(onChange)
  })

  onUnmounted(() => {
    if (!mql) return
    if (mql.removeEventListener) mql.removeEventListener('change', onChange)
    else mql.removeListener(onChange)
  })

  return matches
}
