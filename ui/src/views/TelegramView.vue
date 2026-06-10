<script setup lang="ts">
import { ref, onMounted } from 'vue'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import relativeTime from 'dayjs/plugin/relativeTime'
import { Send, Eye, EyeOff, Trash2, Plus, CheckCircle2, XCircle, Power, PowerOff, Loader2 } from 'lucide-vue-next'
import { telegramApi } from '@/api/telegram'
import { useUIStore } from '@/stores/ui'
import EmptyState from '@/components/EmptyState.vue'
import Skeleton from '@/components/Skeleton.vue'
import type { TelegramConfig, TelegramSubscriber } from '@/api/types'

dayjs.extend(utc)
dayjs.extend(relativeTime)

const ui = useUIStore()
const config = ref<TelegramConfig | null>(null)
const subs = ref<TelegramSubscriber[]>([])
const token = ref('')
const showToken = ref(false)
const testing = ref(false)
const connecting = ref(false)
const sending = ref(false)
const connectionResult = ref<{ ok: boolean; username?: string; error?: string } | null>(null)
const loading = ref(true)

const newChatId = ref('')
const newLabel = ref('')
const adding = ref(false)

onMounted(async () => {
  try {
    const [c, s] = await Promise.all([telegramApi.getConfig(), telegramApi.listSubscribers()])
    config.value = c
    subs.value = s
  } finally {
    loading.value = false
  }
})

async function save() {
  if (!config.value) return
  try {
    const updated = await telegramApi.updateConfig({
      bot_token: token.value || undefined,
      enabled: config.value.enabled,
      notify_on_flag_only: config.value.notify_on_flag_only,
      throttle_ms: config.value.throttle_ms,
    })
    config.value = updated
    token.value = ''
    ui.showToast('Telegram settings saved', 'success')
  } catch (e) {
    ui.showToast('Failed to save', 'error')
  }
}

async function connect() {
  if (!token.value) {
    ui.showToast('Enter a bot token first', 'info')
    return
  }
  connecting.value = true
  connectionResult.value = null
  try {
    await telegramApi.updateConfig({ bot_token: token.value })
    const res = await telegramApi.testConnection()
    connectionResult.value = res
    if (res.ok) ui.showToast('Bot connected', 'success')
  } catch (e) {
    connectionResult.value = { ok: false, error: String(e) }
  } finally {
    connecting.value = false
  }
}

async function sendTest() {
  sending.value = true
  try {
    const results = await telegramApi.testSend()
    const ok = results.filter((r) => r.ok).length
    const fail = results.length - ok
    ui.showToast(`Sent to ${ok}${fail ? `, ${fail} failed` : ''}`, fail ? 'error' : 'success')
  } catch {
    ui.showToast('Test send failed', 'error')
  } finally {
    sending.value = false
  }
}

async function addSub() {
  if (!newChatId.value.trim()) return
  adding.value = true
  try {
    const sub = await telegramApi.addSubscriber(newChatId.value.trim(), newLabel.value.trim() || 'Subscriber')
    subs.value = [sub, ...subs.value]
    newChatId.value = ''
    newLabel.value = ''
    ui.showToast('Subscriber added', 'success')
  } catch {
    ui.showToast('Failed to add', 'error')
  } finally {
    adding.value = false
  }
}

async function toggleSub(s: TelegramSubscriber) {
  try {
    const updated = await telegramApi.updateSubscriber(s.id, { enabled: !s.enabled })
    subs.value = subs.value.map((x) => (x.id === s.id ? updated : x))
  } catch {
    ui.showToast('Failed to update', 'error')
  }
}

async function removeSub(s: TelegramSubscriber) {
  if (!confirm(`Remove ${s.label}?`)) return
  try {
    await telegramApi.removeSubscriber(s.id)
    subs.value = subs.value.filter((x) => x.id !== s.id)
    ui.showToast('Removed', 'success')
  } catch {
    ui.showToast('Failed to remove', 'error')
  }
}

const fmt = (s?: string | null) => (s ? dayjs.utc(s).fromNow() : '—')
</script>

<template>
  <div class="bg-noise min-h-full bg-ink-950 p-4 lg:p-6">
    <div class="mx-auto flex max-w-4xl flex-col gap-4">
      <header>
        <h1 class="display text-2xl font-semibold text-thaana-text">Telegram</h1>
        <p class="mt-1 text-sm text-mid">Push flagged items to your chats and channels.</p>
      </header>

      <section v-if="loading" class="flex flex-col gap-2">
        <Skeleton height="200px" />
        <Skeleton height="120px" />
      </section>

      <template v-else-if="config">
        <section class="card flex flex-col gap-4 p-4 lg:p-5">
          <header class="flex items-center justify-between">
            <h2 class="display text-sm font-semibold text-thaana-text">Bot configuration</h2>
            <button
              type="button"
              role="switch"
              :aria-checked="config.enabled"
              class="flex items-center gap-2 rounded-md border border-ink-700 bg-ink-800 px-2 py-1 text-[10px] uppercase tracking-wider text-mid transition-colors hover:text-thaana-text"
              :class="config.enabled ? 'border-signal-green/30 text-signal-green' : ''"
              @click="config.enabled = !config.enabled; save()"
            >
              <component :is="config.enabled ? Power : PowerOff" :size="11" :stroke-width="2" />
              {{ config.enabled ? 'Enabled' : 'Disabled' }}
            </button>
          </header>

          <div class="flex flex-col gap-1.5">
            <label class="text-[10px] font-medium uppercase tracking-wider text-dim">Bot token</label>
            <div class="relative">
              <input
                v-model="token"
                :type="showToken ? 'text' : 'password'"
                class="input-base pr-20 font-mono text-xs"
                :placeholder="config.has_token ? '•••••• (saved)' : '123456:ABC-DEF...'"
              />
              <div class="absolute right-1 top-1/2 flex -translate-y-1/2 items-center gap-1">
                <button type="button" class="btn-icon h-7 w-7" @click="showToken = !showToken">
                  <component :is="showToken ? EyeOff : Eye" :size="13" :stroke-width="1.5" />
                </button>
              </div>
            </div>
            <p v-if="config.has_token" class="text-[10px] text-mid">
              Token is saved. Leave blank to keep current; enter a new value to replace.
              <span v-if="config.bot_username" class="text-thaana-text">@{{ config.bot_username }}</span>
            </p>
          </div>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <label class="flex cursor-pointer items-center gap-2 text-sm text-mid">
              <input
                type="checkbox"
                v-model="config.notify_on_flag_only"
                class="h-4 w-4 rounded-sm border-ink-700 bg-transparent accent-violet-400"
                @change="save()"
              />
              Notify only on flagged items
            </label>
            <label class="flex flex-col gap-1">
              <div class="text-[10px] font-medium uppercase tracking-wider text-dim">Throttle (ms)</div>
              <input
                type="number"
                v-model.number="config.throttle_ms"
                min="0"
                step="100"
                class="input-base"
                @change="save()"
              />
            </label>
          </div>

          <div class="flex flex-wrap items-center gap-2 border-t border-line pt-3">
            <button class="btn-ghost" :disabled="connecting" @click="connect">
              <Loader2 v-if="connecting" :size="13" :stroke-width="2" class="animate-spin" />
              <CheckCircle2 v-else :size="13" :stroke-width="1.75" />
              Test connection
            </button>
            <button class="btn-primary" :disabled="sending" @click="sendTest">
              <Send v-if="!sending" :size="13" :stroke-width="1.75" />
              <Loader2 v-else :size="13" :stroke-width="2" class="animate-spin" />
              Send test message
            </button>
            <div v-if="connectionResult" class="flex items-center gap-1.5 text-[11px]" :class="connectionResult.ok ? 'text-signal-green' : 'text-signal-rose'">
              <component :is="connectionResult.ok ? CheckCircle2 : XCircle" :size="12" :stroke-width="2" />
              <span v-if="connectionResult.ok">@{{ connectionResult.username || 'bot' }} is reachable</span>
              <span v-else>{{ connectionResult.error || 'Connection failed' }}</span>
            </div>
          </div>
        </section>

        <section class="card flex flex-col gap-3 p-4 lg:p-5">
          <header>
            <h2 class="display text-sm font-semibold text-thaana-text">Subscribers</h2>
            <p class="mt-0.5 text-xs text-mid">Chats that will receive notifications.</p>
          </header>

          <div class="grid grid-cols-1 gap-2 sm:grid-cols-[1fr_1fr_auto]">
            <input
              v-model="newChatId"
              type="text"
              inputmode="numeric"
              placeholder="Chat ID"
              class="input-base font-mono text-xs"
            />
            <input
              v-model="newLabel"
              type="text"
              placeholder="Label (optional)"
              class="input-base"
            />
            <button class="btn-primary" :disabled="adding || !newChatId" @click="addSub">
              <Plus :size="14" :stroke-width="2" />
              Add
            </button>
          </div>

          <EmptyState
            v-if="!subs.length"
            variant="subscribers"
            title="No subscribers"
            subtitle="Add a chat ID to start receiving notifications."
          />

          <ul v-else class="flex flex-col divide-y divide-line">
            <li v-for="s in subs" :key="s.id" class="flex items-center gap-3 py-3 text-sm">
              <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-ink-800 font-mono text-[10px] text-mid">
                {{ s.chat_id.slice(0, 4) }}
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <span class="display truncate text-thaana-text">{{ s.label }}</span>
                  <span class="font-mono text-[10px] text-dim">{{ s.chat_id }}</span>
                </div>
                <div class="mt-0.5 flex items-center gap-2 text-[10px] text-mid">
                  <span>added {{ fmt(s.added_at) }}</span>
                  <span v-if="s.last_delivery_at">· last {{ fmt(s.last_delivery_at) }}</span>
                  <span
                    v-if="s.last_delivery_status"
                    class="inline-flex items-center gap-1"
                    :class="s.last_delivery_status === 'ok' ? 'text-signal-green' : 'text-signal-rose'"
                  >
                    ·
                    <component :is="s.last_delivery_status === 'ok' ? CheckCircle2 : XCircle" :size="10" :stroke-width="2" />
                    {{ s.last_delivery_status }}
                  </span>
                </div>
              </div>
              <div class="flex items-center gap-1.5">
                <button class="btn-icon" @click="toggleSub(s)">
                  <component :is="s.enabled ? Power : PowerOff" :size="13" :stroke-width="1.5" />
                </button>
                <button class="btn-icon text-signal-rose" @click="removeSub(s)">
                  <Trash2 :size="13" :stroke-width="1.5" />
                </button>
              </div>
            </li>
          </ul>
        </section>
      </template>
    </div>
  </div>
</template>
