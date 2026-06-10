<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import TopBar from '@/components/TopBar.vue'
import Sidebar from '@/components/Sidebar.vue'
import MobileDrawer from '@/components/MobileDrawer.vue'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { logsApi } from '@/api/logs'
import { useMediaQuery } from '@/composables/useMediaQuery'
import { useSSE } from '@/composables/useSSE'
import type { LogEntry } from '@/api/types'

const auth = useAuthStore()
const ui = useUIStore()
const router = useRouter()
const route = useRoute()
const isDesktop = useMediaQuery('(min-width: 1024px)')

const isPublic = computed(() => Boolean(route.meta.public))

useSSE<LogEntry>('/api/logs/stream')

async function bootstrap() {
  if (auth.access) {
    try {
      await auth.fetchMe()
    } catch {
      // ignore
    }
  }
  try {
    const recent = await logsApi.recent(200)
    ui.seedLogs(recent)
  } catch {
    // ignore
  }
}

onMounted(bootstrap)

router.beforeEach((to) => {
  if (to.meta.public) return true
  if (!auth.isAuthenticated) {
    return { name: 'login', query: { next: to.fullPath } }
  }
  return true
})
</script>

<template>
  <div class="bg-noise flex h-full min-h-screen flex-col bg-ink-950 text-thaana-text">
    <template v-if="!isPublic">
      <TopBar />
      <div class="flex flex-1 items-stretch">
        <Sidebar />
        <main class="min-w-0 flex-1 overflow-x-hidden">
          <router-view v-slot="{ Component, route: r }">
            <Transition
              mode="out-in"
              enter-active-class="animate-slide-up"
              leave-active-class="transition-opacity duration-100"
              leave-to-class="opacity-0"
            >
              <component :is="Component" :key="r.fullPath" />
            </Transition>
          </router-view>
        </main>
      </div>
      <MobileDrawer />
    </template>
    <template v-else>
      <router-view v-slot="{ Component, route: r }">
        <Transition
          mode="out-in"
          enter-active-class="animate-fade-in"
          leave-active-class="transition-opacity duration-100"
          leave-to-class="opacity-0"
        >
          <component :is="Component" :key="r.fullPath" />
        </Transition>
      </router-view>
    </template>

    <Transition
      enter-active-class="transition duration-200 ease-out-expo"
      leave-active-class="transition duration-200 ease-out-expo"
      enter-from-class="translate-y-2 opacity-0"
      leave-to-class="translate-y-2 opacity-0"
    >
      <div
        v-if="ui.toast"
        class="pointer-events-none fixed inset-x-0 bottom-4 z-[60] flex justify-center px-4 safe-pb"
      >
        <div
          class="rounded-lg border px-3.5 py-2 text-xs backdrop-blur-md"
          :class="[
            ui.toast.kind === 'error'
              ? 'border-signal-rose/40 bg-ink-900/95 text-signal-rose'
              : ui.toast.kind === 'success'
                ? 'border-signal-green/40 bg-ink-900/95 text-signal-green'
                : 'border-ink-700 bg-ink-900/95 text-thaana-text',
          ]"
        >
          {{ ui.toast.msg }}
        </div>
      </div>
    </Transition>
  </div>
</template>
