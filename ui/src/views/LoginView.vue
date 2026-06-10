<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { LogIn, Eye, EyeOff, ShieldCheck } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { api, getErrorMessage } from '@/api/client'
import { setupApi } from '@/api/setup'
import ThaanaText from '@/components/ThaanaText.vue'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const username = ref(import.meta.env.VITE_ADMIN_USERNAME || '')
const password = ref(import.meta.env.VITE_ADMIN_PASSWORD || '')
const showPwd = ref(false)
const submitting = ref(false)
const localError = ref<string | null>(null)
const debugOutput = ref<string>('')

async function runDiagnostic() {
  debugOutput.value = '...'
  try {
    const echo = await api.get('/_echo')
    debugOutput.value = JSON.stringify(echo.data, null, 2)
  } catch (e) {
    debugOutput.value = 'ERR: ' + getErrorMessage(e)
  }
}

onMounted(async () => {
  try {
    const info = await setupApi.info()
    if (info.setup_required) {
      router.replace('/setup')
      return
    }
  } catch {
    // ignore
  }
  if (auth.isAuthenticated) {
    const next = (route.query.next as string) || '/dashboard'
    router.replace(next)
  }
})

async function onSubmit() {
  if (submitting.value) return
  submitting.value = true
  localError.value = null
  try {
    await auth.login(username.value.trim(), password.value)
    const next = (route.query.next as string) || '/dashboard'
    router.push(next)
  } catch (e) {
    localError.value = getErrorMessage(e)
  } finally {
    submitting.value = false
  }
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Enter') onSubmit()
}
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="bg-noise flex min-h-screen bg-ink-950">
    <div class="relative hidden w-1/2 overflow-hidden md:block">
      <div
        class="absolute inset-0 animate-aurora-loop"
        style="
          background:
            radial-gradient(60% 60% at 30% 30%, rgba(255, 107, 157, 0.4), transparent 60%),
            radial-gradient(60% 60% at 70% 70%, rgba(94, 196, 255, 0.4), transparent 60%),
            radial-gradient(40% 40% at 50% 80%, rgba(79, 240, 200, 0.3), transparent 60%),
            #0b0e14;
        "
      />
      <div class="relative flex h-full flex-col justify-between p-10">
        <div class="display text-3xl font-bold text-aurora">Azx</div>
        <div class="max-w-sm">
          <h1 class="display text-2xl font-semibold leading-tight text-thaana-text">
            Every government notice, in one place.
          </h1>
          <p class="mt-3 text-sm text-mid">
            Watch tenders, vacancies, and announcements from across the Maldives. Get flagged the moment something matters.
          </p>
          <ThaanaText
            as="p"
            text="ހުޅުވާ ނިޔަލުން، އެކައުންމެ ތަރުޤީޤު އަދި ޚިދުމަތެއް."
            class="mt-4 block text-sm text-dim"
          />
        </div>
        <div class="flex items-center gap-2 text-[11px] text-dim">
          <ShieldCheck :size="12" :stroke-width="1.5" />
          JWT-secured · self-hosted
        </div>
      </div>
    </div>

    <div class="flex w-full flex-col items-center justify-center px-6 py-10 md:w-1/2 md:px-12">
      <div class="w-full max-w-sm">
        <div class="mb-8 md:hidden">
          <div class="display text-3xl font-bold text-aurora">Azx</div>
        </div>
        <h2 class="display text-xl font-semibold text-thaana-text">Welcome back</h2>
        <p class="mt-1 text-sm text-mid">Sign in to your Scrap Bot control panel.</p>

        <form class="mt-8 flex flex-col gap-4" @submit.prevent="onSubmit">
          <div class="flex flex-col gap-1.5">
            <label class="text-[11px] font-medium uppercase tracking-wider text-dim">Username</label>
            <input
              v-model="username"
              type="text"
              autocomplete="username"
              required
              autofocus
              class="input-base"
              placeholder="admin"
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="text-[11px] font-medium uppercase tracking-wider text-dim">Password</label>
            <div class="relative">
              <input
                v-model="password"
                :type="showPwd ? 'text' : 'password'"
                autocomplete="current-password"
                required
                class="input-base pr-10"
                placeholder="••••••••"
              />
              <button
                type="button"
                class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-dim transition-colors hover:text-mid"
                :aria-label="showPwd ? 'Hide password' : 'Show password'"
                @click="showPwd = !showPwd"
              >
                <component :is="showPwd ? EyeOff : Eye" :size="14" :stroke-width="1.5" />
              </button>
            </div>
          </div>

          <div
            v-if="localError || auth.error"
            class="rounded-lg border border-signal-rose/30 bg-signal-rose/10 px-3 py-2 text-xs text-signal-rose"
          >
            {{ localError || auth.error }}
          </div>

          <button
            type="submit"
            class="btn-primary mt-2 h-10"
            :disabled="submitting"
          >
            <LogIn :size="14" :stroke-width="2" />
            <span>{{ submitting ? 'Signing in…' : 'Sign in' }}</span>
          </button>
        </form>

        <div class="mt-6 text-center text-xs text-dim">
          First time?
          <router-link to="/setup" class="text-violet-400 hover:underline">Set up admin</router-link>
        </div>

        <details class="mt-6">
          <summary class="cursor-pointer text-[11px] text-dim hover:text-mid">Diagnostic</summary>
          <div class="mt-2 space-y-2">
            <button
              type="button"
              class="rounded-lg border border-ink-700 bg-ink-800/60 px-3 py-1.5 text-[11px] text-mid hover:bg-ink-700"
              @click="runDiagnostic"
            >
              Run /api/_echo
            </button>
            <pre v-if="debugOutput" class="max-h-48 overflow-auto rounded-lg border border-ink-700 bg-ink-900 p-2 text-[10px] text-thaana-text">{{ debugOutput }}</pre>
          </div>
        </details>
      </div>
    </div>
  </div>
</template>
