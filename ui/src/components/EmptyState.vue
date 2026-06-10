<script setup lang="ts">
defineProps<{
  variant: 'sources' | 'items' | 'flags' | 'subscribers'
  title: string
  subtitle?: string
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center px-6 py-12 text-center">
    <div class="mb-5 h-20 w-20">
      <svg v-if="variant === 'sources'" viewBox="0 0 80 80" class="h-full w-full">
        <defs>
          <linearGradient id="es-s" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stop-color="#FF6B9D" />
            <stop offset="50%" stop-color="#B968FF" />
            <stop offset="100%" stop-color="#5EC4FF" />
          </linearGradient>
        </defs>
        <circle cx="40" cy="40" r="28" fill="none" stroke="url(#es-s)" stroke-width="2" opacity="0.3" />
        <circle cx="40" cy="40" r="18" fill="none" stroke="url(#es-s)" stroke-width="2" opacity="0.5" />
        <circle cx="40" cy="40" r="3" fill="url(#es-s)" />
        <path d="M40 4 L40 14 M40 66 L40 76 M4 40 L14 40 M66 40 L76 40" stroke="url(#es-s)" stroke-width="2" stroke-linecap="round" opacity="0.5" />
      </svg>
      <svg v-else-if="variant === 'items'" viewBox="0 0 80 80" class="h-full w-full">
        <defs>
          <linearGradient id="es-i" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stop-color="#B968FF" />
            <stop offset="100%" stop-color="#4FF0C8" />
          </linearGradient>
        </defs>
        <rect x="14" y="18" width="52" height="44" rx="4" fill="none" stroke="url(#es-i)" stroke-width="2" opacity="0.4" />
        <line x1="22" y1="30" x2="58" y2="30" stroke="url(#es-i)" stroke-width="2" opacity="0.6" />
        <line x1="22" y1="40" x2="48" y2="40" stroke="url(#es-i)" stroke-width="2" opacity="0.4" />
        <line x1="22" y1="50" x2="54" y2="50" stroke="url(#es-i)" stroke-width="2" opacity="0.4" />
      </svg>
      <svg v-else-if="variant === 'flags'" viewBox="0 0 80 80" class="h-full w-full">
        <defs>
          <linearGradient id="es-f" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stop-color="#FF6B9D" />
            <stop offset="100%" stop-color="#FFB454" />
          </linearGradient>
        </defs>
        <path d="M40 8 L48 28 L70 30 L52 44 L58 66 L40 54 L22 66 L28 44 L10 30 L32 28 Z" fill="none" stroke="url(#es-f)" stroke-width="2" opacity="0.5" />
      </svg>
      <svg v-else viewBox="0 0 80 80" class="h-full w-full">
        <defs>
          <linearGradient id="es-u" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stop-color="#5EC4FF" />
            <stop offset="100%" stop-color="#4FF0C8" />
          </linearGradient>
        </defs>
        <circle cx="40" cy="30" r="10" fill="none" stroke="url(#es-u)" stroke-width="2" opacity="0.5" />
        <path d="M20 60 C20 50 28 46 40 46 C52 46 60 50 60 60" fill="none" stroke="url(#es-u)" stroke-width="2" stroke-linecap="round" opacity="0.5" />
      </svg>
    </div>
    <h3 class="display text-sm font-semibold text-thaana-text">{{ title }}</h3>
    <p v-if="subtitle" class="mt-1 max-w-xs text-xs text-mid">{{ subtitle }}</p>
    <div class="mt-4">
      <slot />
    </div>
  </div>
</template>
