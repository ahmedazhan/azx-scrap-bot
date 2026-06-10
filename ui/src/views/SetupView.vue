<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { UserPlus, Eye, EyeOff, KeyRound, ShieldCheck } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { getErrorMessage } from '@/api/client'
import { setupApi } from '@/api/setup'
import ThaanaText from '@/components/ThaanaText.vue'

const auth = useAuthStore()
const router = useRouter()

const username = ref(import.meta.env.VITE_ADMIN_USERNAME || 'azhan')
const password = ref(import.meta.env.VITE_ADMIN_PASSWORD || '')
const confirm = ref(import.meta.env.VITE_ADMIN_PASSWORD || '')
const token = ref(import.meta.env.VITE_SETUP_TOKEN || '')
const tokenFromEnv = computed(() => Boolean(import.meta.env.VITE_SETUP_TOKEN))
const showPwd = ref(false)
const submitting = ref(false)
const localError = ref<string | null>(null)

onMounted(async () => {
  try {
    const info = await setupApi.info()
    if (!info.setup_required) {
      router.replace('/login')
    }
  } catch {
    // ignore
  }
})

async function onSubmit() {
  localError.value = null
  if (password.value.length < 8) {
    localError.value = 'Password must be at least 8 characters'
    return
  }
  if (password.value !== confirm.value) {
    localError.value = "Passwords don't match"
    return
  }
  submitting.value = true
  try {
    await auth.setup(username.value.trim(), password.value, token.value.trim())
    router.push('/dashboard')
  } catch (e) {
    localError.value = getErrorMessage(e)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="bg-noise flex min-h-screen bg-ink-950">
    <div class="relative hidden w-1/2 overflow-hidden md:block">
      <div
        class="absolute inset-0 animate-aurora-loop"
        style="
          background:
            radial-gradient(60% 60% at 70% 30%, rgba(79, 240, 200, 0.35), transparent 60%),
            radial-gradient(60% 60% at 30% 70%, rgba(185, 104, 255, 0.35), transparent 60%),
            radial-gradient(50% 50% at 60% 60%, rgba(255, 107, 157, 0.3), transparent 60%),
            #0b0e14;
        "
      />
      <div class="relative flex h-full flex-col justify-between p-10">
        <div class="display text-3xl font-bold text-aurora">Azx</div>
        <div class="max-w-sm">
          <h1 class="display text-2xl font-semibold leading-tight text-thaana-text">
            First time? Let's get you set up.
          </h1>
          <p class="mt-3 text-sm text-mid">
            Create the admin account for your Scrap Bot instance. You'll need a setup token from your server config.
          </p>
          <ThaanaText as="p" text="އިސްޓިޤާމާއި ޤާނޫން ދިޔުމަށް ޚިދުމަތެއް." class="mt-4 block text-sm text-dim" />
        </div>
        <div class="flex items-center gap-2 text-[11px] text-dim">
          <ShieldCheck :size="12" :stroke-width="1.5" />
          One-time admin onboarding
        </div>
      </div>
    </div>

    <div class="flex w-full flex-col items-center justify-center px-6 py-10 md:w-1/2 md:px-12">
      <div class="w-full max-w-sm">
        <div class="mb-8 md:hidden">
          <div class="display text-3xl font-bold text-aurora">Azx</div>
        </div>
        <h2 class="display text-xl font-semibold text-thaana-text">Create admin account</h2>
        <p class="mt-1 text-sm text-mid">Use a strong password — you'll need it to sign in.</p>

        <form class="mt-8 flex flex-col gap-4" @submit.prevent="onSubmit">
          <div class="flex flex-col gap-1.5">
            <label class="text-[11px] font-medium uppercase tracking-wider text-dim">Setup token</label>
            <div class="relative">
              <KeyRound
                :size="14"
                :stroke-width="1.5"
                class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-dim"
              />
              <input
                v-model="token"
                type="text"
                required
                class="input-base pl-8 font-mono text-xs"
                placeholder="paste setup token"
              />
            </div>
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="text-[11px] font-medium uppercase tracking-wider text-dim">Username</label>
            <input
              v-model="username"
              type="text"
              autocomplete="username"
              required
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
                autocomplete="new-password"
                required
                class="input-base pr-10"
                placeholder="8+ characters"
              />
              <button
                type="button"
                class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-dim transition-colors hover:text-mid"
                @click="showPwd = !showPwd"
              >
                <component :is="showPwd ? EyeOff : Eye" :size="14" :stroke-width="1.5" />
              </button>
            </div>
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="text-[11px] font-medium uppercase tracking-wider text-dim">Confirm password</label>
            <input
              v-model="confirm"
              :type="showPwd ? 'text' : 'password'"
              autocomplete="new-password"
              required
              class="input-base"
              placeholder="repeat password"
            />
          </div>

          <div
            v-if="tokenFromEnv"
            class="rounded-lg border border-signal-green/30 bg-signal-green/10 px-3 py-2 text-xs text-signal-green"
          >
            Token pre-filled from <code class="font-mono">VITE_SETUP_TOKEN</code>. Just set a username and password.
          </div>

          <div
            v-if="localError"
            class="rounded-lg border border-signal-rose/30 bg-signal-rose/10 px-3 py-2 text-xs text-signal-rose"
          >
            {{ localError }}
          </div>

          <button
            type="submit"
            class="btn-primary mt-2 h-10"
            :disabled="submitting"
          >
            <UserPlus :size="14" :stroke-width="2" />
            <span>{{ submitting ? 'Creating…' : 'Create admin' }}</span>
          </button>
        </form>

        <div class="mt-6 text-center text-xs text-dim">
          Already have an account?
          <router-link to="/login" class="text-violet-400 hover:underline">Sign in</router-link>
        </div>
      </div>
    </div>
  </div>
</template>
