<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Sun, Moon, LogOut, KeyRound, Save, Smartphone, Layers } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { accountApi } from '@/api/account'
import { getErrorMessage } from '@/api/client'
import type { User } from '@/api/types'

const auth = useAuthStore()
const ui = useUIStore()
const router = useRouter()

const user = ref<User | null>(null)
const loading = ref(true)

const oldPwd = ref('')
const newPwd = ref('')
const confirmPwd = ref('')
const saving = ref(false)
const message = ref<{ kind: 'ok' | 'err'; text: string } | null>(null)

onMounted(async () => {
  try {
    user.value = await accountApi.get()
    if (user.value) {
      if (user.value.theme) ui.setTheme(user.value.theme)
      if (user.value.filter_mode) ui.setFilterMode(user.value.filter_mode)
    }
  } finally {
    loading.value = false
  }
})

const version = import.meta.env.VITE_APP_VERSION || '0.1.0'

async function saveAccount() {
  if (!user.value) return
  saving.value = true
  message.value = null
  try {
    const updated = await accountApi.update({
      theme: ui.theme,
      filter_mode: ui.filterMode,
      pull_to_refresh: user.value.pull_to_refresh,
    })
    user.value = updated
    auth.setUser(updated)
    ui.showToast('Preferences saved', 'success')
  } catch (e) {
    message.value = { kind: 'err', text: getErrorMessage(e) }
  } finally {
    saving.value = false
  }
}

async function changePassword() {
  message.value = null
  if (newPwd.value.length < 8) {
    message.value = { kind: 'err', text: 'New password must be at least 8 characters' }
    return
  }
  if (newPwd.value !== confirmPwd.value) {
    message.value = { kind: 'err', text: "Passwords don't match" }
    return
  }
  saving.value = true
  try {
    await auth.changePassword(oldPwd.value, newPwd.value)
    message.value = { kind: 'ok', text: 'Password changed' }
    oldPwd.value = ''
    newPwd.value = ''
    confirmPwd.value = ''
  } catch (e) {
    message.value = { kind: 'err', text: getErrorMessage(e) }
  } finally {
    saving.value = false
  }
}

function logout() {
  auth.logout()
  router.push('/login')
}

const initials = computed(() => (user.value?.username || 'A').slice(0, 1).toUpperCase())
</script>

<template>
  <div class="bg-noise min-h-full bg-ink-950 p-4 lg:p-6">
    <div class="mx-auto flex max-w-2xl flex-col gap-4">
      <header>
        <h1 class="display text-2xl font-semibold text-thaana-text">Account</h1>
        <p class="mt-1 text-sm text-mid">Manage your session, password, and preferences.</p>
      </header>

      <section v-if="loading" class="flex flex-col gap-2">
        <Skeleton height="120px" />
        <Skeleton height="240px" />
      </section>

      <template v-else-if="user">
        <section class="card flex items-center gap-4 p-4">
          <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-aurora text-base font-semibold text-white">
            {{ initials }}
          </div>
          <div class="min-w-0 flex-1">
            <div class="display text-sm font-semibold text-thaana-text">{{ user.username }}</div>
            <div class="text-[11px] text-dim">Administrator · v{{ version }}</div>
          </div>
          <button class="btn-danger" @click="logout">
            <LogOut :size="13" :stroke-width="1.75" />
            Logout
          </button>
        </section>

        <section class="card flex flex-col gap-3 p-4">
          <h2 class="display text-sm font-semibold text-thaana-text">Change password</h2>
          <form class="flex flex-col gap-2.5" @submit.prevent="changePassword">
            <input v-model="oldPwd" type="password" placeholder="Current password" class="input-base" required />
            <input v-model="newPwd" type="password" placeholder="New password (8+ chars)" class="input-base" required />
            <input v-model="confirmPwd" type="password" placeholder="Confirm new password" class="input-base" required />
            <div v-if="message" class="rounded-md border px-3 py-2 text-xs"
                 :class="message.kind === 'ok' ? 'border-signal-green/30 bg-signal-green/10 text-signal-green' : 'border-signal-rose/30 bg-signal-rose/10 text-signal-rose'">
              {{ message.text }}
            </div>
            <div class="flex justify-end">
              <button type="submit" class="btn-primary" :disabled="saving">
                <KeyRound :size="13" :stroke-width="1.75" />
                Update password
              </button>
            </div>
          </form>
        </section>

        <section class="card flex flex-col gap-3 p-4">
          <h2 class="display text-sm font-semibold text-thaana-text">Preferences</h2>

          <div class="flex items-center justify-between gap-2 rounded-lg border border-ink-700 bg-ink-800/40 p-3">
            <div class="flex items-center gap-2">
              <component :is="ui.theme === 'dark' ? Moon : Sun" :size="14" :stroke-width="1.5" class="text-mid" />
              <div>
                <div class="text-sm text-thaana-text">Theme</div>
                <div class="text-[10px] text-dim">Currently {{ ui.theme }}</div>
              </div>
            </div>
            <div class="flex rounded-md border border-ink-700 bg-ink-900 p-0.5">
              <button
                v-for="t in ['dark', 'light'] as const"
                :key="t"
                class="rounded px-2.5 py-1 text-[11px] capitalize transition-colors"
                :class="ui.theme === t ? 'bg-ink-700 text-thaana-text' : 'text-mid hover:text-thaana-text'"
                @click="ui.setTheme(t)"
              >
                {{ t }}
              </button>
            </div>
          </div>

          <div class="flex items-center justify-between gap-2 rounded-lg border border-ink-700 bg-ink-800/40 p-3">
            <div class="flex items-center gap-2">
              <Layers :size="14" :stroke-width="1.5" class="text-mid" />
              <div>
                <div class="text-sm text-thaana-text">Filter mode</div>
                <div class="text-[10px] text-dim">How filters open in the items view</div>
              </div>
            </div>
            <div class="flex rounded-md border border-ink-700 bg-ink-900 p-0.5">
              <button
                v-for="m in ['sheet', 'inline'] as const"
                :key="m"
                class="rounded px-2.5 py-1 text-[11px] capitalize transition-colors"
                :class="ui.filterMode === m ? 'bg-ink-700 text-thaana-text' : 'text-mid hover:text-thaana-text'"
                @click="ui.setFilterMode(m)"
              >
                {{ m }}
              </button>
            </div>
          </div>

          <label class="flex cursor-pointer items-center justify-between gap-2 rounded-lg border border-ink-700 bg-ink-800/40 p-3">
            <div class="flex items-center gap-2">
              <Smartphone :size="14" :stroke-width="1.5" class="text-mid" />
              <div>
                <div class="text-sm text-thaana-text">Pull to refresh</div>
                <div class="text-[10px] text-dim">Swipe down on items to run all sources</div>
              </div>
            </div>
            <button
              type="button"
              role="switch"
              :aria-checked="user.pull_to_refresh ?? false"
              class="relative h-5 w-9 rounded-full transition-colors duration-150"
              :class="user.pull_to_refresh ? 'bg-aurora' : 'bg-ink-700'"
              @click="user.pull_to_refresh = !user.pull_to_refresh"
            >
              <span
                class="absolute top-0.5 h-4 w-4 rounded-full bg-white transition-transform duration-150 ease-out-expo"
                :class="user.pull_to_refresh ? 'translate-x-4' : 'translate-x-0.5'"
              />
            </button>
          </label>

          <div class="flex justify-end">
            <button class="btn-primary" :disabled="saving" @click="saveAccount">
              <Save :size="13" :stroke-width="1.75" />
              Save preferences
            </button>
          </div>
        </section>
      </template>
    </div>
  </div>
</template>
