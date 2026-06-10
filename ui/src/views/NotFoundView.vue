<script setup lang="ts">
import { ArrowLeft } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import EmptyState from '@/components/EmptyState.vue'

const router = useRouter()
</script>

<template>
  <div class="bg-noise flex min-h-screen items-center justify-center bg-ink-950 p-6">
    <div class="text-center">
      <EmptyState variant="items" title="404 — not found" subtitle="The page you're looking for doesn't exist.">
        <button class="btn-ghost mt-2" @click="router.push('/dashboard')">
          <ArrowLeft :size="14" :stroke-width="1.5" />
          Back to dashboard
        </button>
      </EmptyState>
    </div>
  </div>
</template>
