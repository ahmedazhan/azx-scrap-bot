<script setup lang="ts">
defineProps<{ visible?: boolean }>()
</script>

<template>
  <div
    v-if="visible !== false"
    class="sticky bottom-0 z-30 mt-4 -mx-4 border-t border-ink-700 bg-ink-900/95 px-4 py-3 backdrop-blur-md safe-pb lg:hidden"
  >
    <slot />
  </div>
</template>
