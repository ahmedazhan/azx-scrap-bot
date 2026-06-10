<script setup lang="ts">
import { watch, onUnmounted } from 'vue'
import { X } from 'lucide-vue-next'

const props = defineProps<{ open: boolean; title?: string; width?: string }>()
const emit = defineEmits<{ (e: 'close'): void }>()

watch(
  () => props.open,
  (v) => {
    document.body.style.overflow = v ? 'hidden' : ''
  },
)
onUnmounted(() => {
  document.body.style.overflow = ''
})

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.open) emit('close')
}
if (typeof window !== 'undefined') {
  window.addEventListener('keydown', onKey)
  onUnmounted(() => window.removeEventListener('keydown', onKey))
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-150 ease-out-expo"
      leave-active-class="transition-opacity duration-150 ease-out-expo"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-40 bg-ink-950/60 backdrop-blur-sm"
        @click="emit('close')"
      />
    </Transition>
    <Transition
      enter-active-class="transition-transform duration-200 ease-out-expo"
      leave-active-class="transition-transform duration-200 ease-out-expo"
      enter-from-class="translate-x-full"
      leave-to-class="translate-x-full"
    >
      <aside
        v-if="open"
        class="fixed right-0 top-0 z-50 flex h-full flex-col border-l border-ink-700 bg-ink-900 shadow-2xl"
        :style="{ width: width ?? 'min(560px, 100vw)' }"
      >
        <div class="flex h-12 shrink-0 items-center justify-between border-b border-ink-700 px-4">
          <h2 class="display text-sm font-semibold text-thaana-text">{{ title || 'Details' }}</h2>
          <button
            class="btn-icon"
            aria-label="Close"
            @click="emit('close')"
          >
            <X :size="16" :stroke-width="1.5" />
          </button>
        </div>
        <div class="flex-1 overflow-y-auto scroll-thin">
          <slot />
        </div>
      </aside>
    </Transition>
  </Teleport>
</template>
