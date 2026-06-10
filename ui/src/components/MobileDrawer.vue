<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  LayoutDashboard,
  List,
  Radio,
  Filter,
  Send,
  User,
  LogOut,
  X,
} from 'lucide-vue-next'
import { useUIStore } from '@/stores/ui'
import { useAuthStore } from '@/stores/auth'
import { useSwipe } from '@/composables/useSwipe'

const ui = useUIStore()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const items = [
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/items', label: 'Items', icon: List },
  { to: '/sources', label: 'Sources', icon: Radio },
  { to: '/rules', label: 'Rules', icon: Filter },
  { to: '/telegram', label: 'Telegram', icon: Send },
  { to: '/account', label: 'Account', icon: User },
]

const drawerEl = ref<HTMLElement | null>(null)
useSwipe(drawerEl, {
  onSwipeLeft: () => {
    if (ui.drawerOpen) ui.setDrawer(false)
  },
})

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && ui.drawerOpen) ui.setDrawer(false)
}
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))

function navigate(to: string) {
  router.push(to)
  ui.setDrawer(false)
}

function logout() {
  auth.logout()
  router.push('/login')
  ui.setDrawer(false)
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
        v-if="ui.drawerOpen"
        class="fixed inset-0 z-40 bg-ink-950/60 backdrop-blur-sm lg:hidden"
        @click="ui.setDrawer(false)"
      />
    </Transition>
    <Transition
      enter-active-class="transition-transform duration-200 ease-out-expo"
      leave-active-class="transition-transform duration-200 ease-out-expo"
      enter-from-class="-translate-x-full"
      leave-to-class="-translate-x-full"
    >
      <aside
        v-if="ui.drawerOpen"
        ref="drawerEl"
        class="fixed inset-y-0 left-0 z-50 flex w-[280px] flex-col border-r border-ink-700 bg-ink-900 shadow-2xl lg:hidden"
      >
        <div class="flex h-12 items-center justify-between border-b border-ink-700 px-4 safe-pt">
          <div class="flex items-center gap-2">
            <div class="display rounded-2xl px-2 py-0.5 text-base font-bold text-aurora">Azx</div>
            <span class="text-[10px] uppercase tracking-widest text-mid">Scrap Bot</span>
          </div>
          <button class="btn-icon" @click="ui.setDrawer(false)">
            <X :size="16" :stroke-width="1.5" />
          </button>
        </div>
        <nav class="flex flex-1 flex-col gap-1 overflow-y-auto scroll-thin p-3">
          <router-link
            v-for="i in items"
            :key="i.to"
            :to="i.to"
            class="nav-item"
            :class="route.path.startsWith(i.to) ? 'active' : ''"
            @click="navigate(i.to)"
          >
            <component :is="i.icon" :size="18" :stroke-width="1.5" class="shrink-0" />
            <span class="truncate">{{ i.label }}</span>
          </router-link>
        </nav>
        <div class="border-t border-ink-700 p-3 safe-pb">
          <div class="mb-2 flex items-center gap-2 rounded-lg bg-ink-800/60 p-2">
            <div class="flex h-8 w-8 items-center justify-center rounded-md bg-aurora text-xs font-semibold text-white">
              {{ (auth.user?.username || 'A').slice(0, 1).toUpperCase() }}
            </div>
            <div class="min-w-0 flex-1">
              <div class="truncate text-xs text-thaana-text">{{ auth.user?.username || '—' }}</div>
              <div class="truncate text-[10px] text-dim">Administrator</div>
            </div>
          </div>
          <button
            class="flex w-full items-center gap-2 rounded-md bg-ink-800/60 px-3 py-2 text-xs text-signal-rose transition-colors hover:bg-ink-800"
            @click="logout"
          >
            <LogOut :size="12" :stroke-width="1.5" />
            Logout
          </button>
        </div>
      </aside>
    </Transition>
  </Teleport>
</template>
