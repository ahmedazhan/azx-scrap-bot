<script setup lang="ts">
defineProps<{ text: string; as?: 'span' | 'p' | 'div' | 'h1' | 'h2' | 'h3' | 'h4' }>()
</script>

<template>
  <component :is="as ?? 'span'" dir="rtl" class="thaana">{{ text }}</component>
</template>
