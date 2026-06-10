<script setup lang="ts">
defineProps<{
  width?: string
  height?: string
  rounded?: string
}>()
</script>

<template>
  <div
    class="skeleton"
    :style="{
      width: width ?? '100%',
      height: height ?? '1rem',
      borderRadius: rounded ?? '0.375rem',
    }"
  />
</template>
