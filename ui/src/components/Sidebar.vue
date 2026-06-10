<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  LayoutDashboard,
  List,
  Radio,
  Filter,
  Send,
  User,
} from 'lucide-vue-next'
import { useUIStore } from '@/stores/ui'

const route = useRoute()
const ui = useUIStore()

const items = [
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/items', label: 'Items', icon: List },
  { to: '/sources', label: 'Sources', icon: Radio },
  { to: '/rules', label: 'Rules', icon: Filter },
  { to: '/telegram', label: 'Telegram', icon: Send },
  { to: '/account', label: 'Account', icon: User },
]

const expanded = ref(false)
</script>

<template>
  <aside
    class="sticky top-12 hidden h-[calc(100vh-3rem)] shrink-0 border-r border-ink-700 bg-ink-900/40 backdrop-blur-sm transition-[width] duration-150 ease-out-expo lg:block"
    :class="expanded ? 'w-[220px]' : 'w-14'"
    @mouseenter="ui.setSidebarExpanded(true); expanded = true"
    @mouseleave="ui.setSidebarExpanded(false); expanded = false"
  >
    <nav class="flex h-full flex-col gap-1 px-2 py-3">
      <router-link
        v-for="i in items"
        :key="i.to"
        :to="i.to"
        class="nav-item"
        :class="[
          route.path.startsWith(i.to) ? 'active' : '',
          !expanded ? 'justify-center px-0' : '',
        ]"
        :title="!expanded ? i.label : undefined"
      >
        <component :is="i.icon" :size="18" :stroke-width="1.5" class="shrink-0" />
        <span
          v-if="expanded"
          class="truncate text-[13px] transition-opacity duration-150"
        >
          {{ i.label }}
        </span>
      </router-link>
    </nav>
  </aside>
</template>
