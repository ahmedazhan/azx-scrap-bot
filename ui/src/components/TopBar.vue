<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Menu, Sun, Moon, LogOut, ChevronDown, Radio } from 'lucide-vue-next'
import RadarMark from './RadarMark.vue'
import { useUIStore } from '@/stores/ui'
import { useAuthStore } from '@/stores/auth'

const ui = useUIStore()
const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const open = ref(false)
const menuRef = ref<HTMLElement | null>(null)

const links = [
  { to: '/dashboard', label: 'Dashboard' },
  { to: '/items', label: 'Items' },
  { to: '/sources', label: 'Sources' },
  { to: '/rules', label: 'Rules' },
  { to: '/telegram', label: 'Telegram' },
]

const active = computed(() => links.find((l) => route.path.startsWith(l.to)))

function toggleTheme() {
  ui.setTheme(ui.theme === 'dark' ? 'light' : 'dark')
}

function logout() {
  auth.logout()
  router.push('/login')
}

function onDocClick(e: MouseEvent) {
  if (!open.value) return
  if (menuRef.value && !menuRef.value.contains(e.target as Node)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))

const version = import.meta.env.VITE_APP_VERSION || '0.1.0'
</script>

<template>
  <header
    class="sticky top-0 z-30 flex h-12 items-center gap-2 border-b border-ink-700 bg-ink-900/80 px-3 backdrop-blur-md safe-pt lg:px-4"
  >
    <button
      class="btn-icon lg:hidden"
      aria-label="Open menu"
      @click="ui.toggleDrawer()"
    >
      <Menu :size="18" :stroke-width="1.5" />
    </button>

    <router-link to="/dashboard" class="flex items-center gap-2">
      <div class="display rounded-2xl px-2 py-0.5 text-base font-bold text-aurora">Azx</div>
      <span class="hidden text-[11px] font-medium uppercase tracking-widest text-mid sm:inline">
        Scrap Bot
      </span>
    </router-link>

    <nav class="ml-4 hidden items-center gap-1 lg:flex">
      <router-link
        v-for="l in links"
        :key="l.to"
        :to="l.to"
        class="nav-item"
        :class="route.path.startsWith(l.to) ? 'active' : ''"
      >
        {{ l.label }}
      </router-link>
    </nav>

    <div class="ml-auto flex items-center gap-1.5">
      <div class="flex items-center gap-1.5 rounded-md bg-ink-800/60 px-2 py-1 text-[11px] text-mid">
        <RadarMark :active="ui.anySourceActive" :size="14" />
        <span class="hidden sm:inline">{{ active?.label || 'Idle' }}</span>
      </div>

      <button
        class="btn-icon"
        :aria-label="ui.theme === 'dark' ? 'Switch to light' : 'Switch to dark'"
        @click="toggleTheme"
      >
        <Sun v-if="ui.theme === 'dark'" :size="16" :stroke-width="1.5" />
        <Moon v-else :size="16" :stroke-width="1.5" />
      </button>

      <div ref="menuRef" class="relative">
        <button
          class="flex h-9 items-center gap-1.5 rounded-lg bg-ink-700 px-2 text-xs font-medium text-thaana-text transition-colors duration-150 ease-out-expo hover:bg-ink-600"
          @click="open = !open"
        >
          <div class="flex h-6 w-6 items-center justify-center rounded-md bg-aurora text-[11px] font-semibold text-white">
            {{ (auth.user?.username || 'A').slice(0, 1).toUpperCase() }}
          </div>
          <span class="hidden sm:inline">{{ auth.user?.username || 'Account' }}</span>
          <ChevronDown :size="12" :stroke-width="2" />
        </button>
        <Transition
          enter-active-class="transition duration-150 ease-out-expo"
          leave-active-class="transition duration-150 ease-out-expo"
          enter-from-class="opacity-0 -translate-y-1"
          leave-to-class="opacity-0 -translate-y-1"
        >
          <div
            v-if="open"
            class="absolute right-0 top-11 w-56 rounded-lg border border-ink-700 bg-ink-900 p-1 shadow-xl"
          >
            <div class="px-3 py-2 text-xs text-mid">
              <div class="text-thaana-text">{{ auth.user?.username || '—' }}</div>
              <div class="text-[10px] text-dim">v{{ version }}</div>
            </div>
            <div class="divider my-1" />
            <button
              class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-xs text-mid transition-colors hover:bg-ink-800 hover:text-thaana-text"
              @click="router.push('/account'); open = false"
            >
              <Radio :size="12" :stroke-width="1.5" /> Account
            </button>
            <button
              class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-xs text-signal-rose transition-colors hover:bg-ink-800"
              @click="logout"
            >
              <LogOut :size="12" :stroke-width="1.5" /> Logout
            </button>
          </div>
        </Transition>
      </div>
    </div>
  </header>
</template>
