<script setup lang="ts">
defineProps<{ active: boolean; size?: number }>()
</script>

<template>
  <svg
    :width="size ?? 16"
    :height="size ?? 16"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
  >
    <defs>
      <linearGradient id="radar-grad" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#FF6B9D" />
        <stop offset="35%" stop-color="#B968FF" />
        <stop offset="70%" stop-color="#5EC4FF" />
        <stop offset="100%" stop-color="#4FF0C8" />
      </linearGradient>
    </defs>
    <circle cx="12" cy="12" r="9" stroke="rgba(255,255,255,0.08)" stroke-width="1.5" />
    <circle cx="12" cy="12" r="5.5" stroke="rgba(255,255,255,0.08)" stroke-width="1.5" />
    <g :class="active ? 'animate-radar-sweep' : ''" style="transform-origin: 12px 12px">
      <path
        d="M12 12 L21 12 A9 9 0 0 0 15.364 3.873 Z"
        fill="url(#radar-grad)"
        :opacity="active ? 0.85 : 0.45"
      />
    </g>
    <circle cx="12" cy="12" r="1.4" fill="url(#radar-grad)" />
  </svg>
</template>
