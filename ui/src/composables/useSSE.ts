import { ref, onUnmounted } from 'vue'
import { useMediaQuery } from './useMediaQuery'

export function useSSE<T = unknown>(url: string) {
  const events = ref<T[]>([]) as { value: T[] }
  const connected = ref(false)
  const isMobile = useMediaQuery('(max-width: 640px)')

  let es: EventSource | null = null
  let retry = 0
  let buffer: T[] = []
  let flushTimer: number | null = null
  let stopped = false

  function flush() {
    if (buffer.length) {
      events.value = [...buffer.reverse(), ...events.value].slice(0, 200)
      buffer = []
    }
    flushTimer = null
  }

  function schedule() {
    if (flushTimer !== null) return
    const delay = isMobile.value ? 250 : 50
    flushTimer = window.setTimeout(flush, delay)
  }

  function connect() {
    if (stopped || typeof window === 'undefined' || typeof EventSource === 'undefined') return
    try {
      es = new EventSource(url, { withCredentials: false } as EventSourceInit)
    } catch {
      scheduleReconnect()
      return
    }
    es.onopen = () => {
      connected.value = true
      retry = 0
    }
    es.onerror = () => {
      connected.value = false
      es?.close()
      es = null
      scheduleReconnect()
    }
    es.onmessage = (msg) => {
      try {
        const data = JSON.parse(msg.data) as T
        buffer.push(data)
        schedule()
      } catch {
        // ignore
      }
    }
  }

  function scheduleReconnect() {
    if (stopped) return
    const delay = Math.min(15000, 500 * Math.pow(2, retry++))
    setTimeout(connect, delay)
  }

  connect()

  onUnmounted(() => {
    stopped = true
    if (flushTimer) clearTimeout(flushTimer)
    if (es) es.close()
  })

  return { events, connected }
}
