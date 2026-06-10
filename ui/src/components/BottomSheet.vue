<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { X } from 'lucide-vue-next'

const props = defineProps<{ open: boolean; title?: string; height?: string }>()
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
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
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
        class="fixed inset-0 z-40 bg-ink-950/60 backdrop-blur-sm lg:hidden"
        @click="emit('close')"
      />
    </Transition>
    <Transition
      enter-active-class="transition-transform duration-200 ease-out-expo"
      leave-active-class="transition-transform duration-200 ease-out-expo"
      enter-from-class="translate-y-full"
      leave-to-class="translate-y-full"
    >
      <div
        v-if="open"
        class="fixed inset-x-0 bottom-0 z-50 flex flex-col rounded-t-2xl border-t border-ink-700 bg-ink-900 shadow-2xl lg:hidden"
        :style="{ maxHeight: height ?? '85vh' }"
      >
        <div class="flex items-center justify-between border-b border-ink-700 px-4 py-3">
          <h2 class="display text-sm font-semibold text-thaana-text">{{ title || 'Filters' }}</h2>
          <button class="btn-icon" aria-label="Close" @click="emit('close')">
            <X :size="16" :stroke-width="1.5" />
          </button>
        </div>
        <div class="flex-1 overflow-y-auto scroll-thin safe-pb">
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
