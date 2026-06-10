<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Plus, Trash2, ToggleLeft, ToggleRight, FlaskConical, Save, X } from 'lucide-vue-next'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import relativeTime from 'dayjs/plugin/relativeTime'
import { rulesApi, type RuleTestResult } from '@/api/rules'
import { useUIStore } from '@/stores/ui'
import EmptyState from '@/components/EmptyState.vue'
import Skeleton from '@/components/Skeleton.vue'
import type { FlagRule } from '@/api/types'

dayjs.extend(utc)
dayjs.extend(relativeTime)

const ui = useUIStore()
const rules = ref<FlagRule[]>([])
const loading = ref(true)
const editing = ref<FlagRule | null>(null)
const isNew = ref(false)
const testResult = ref<RuleTestResult | null>(null)
const testing = ref(false)

onMounted(async () => {
  try {
    rules.value = await rulesApi.list()
  } finally {
    loading.value = false
  }
})

function startNew() {
  editing.value = {
    id: 0,
    name: '',
    pattern: '',
    is_regex: false,
    match_in: 'all',
    case_sensitive: false,
    enabled: true,
  }
  isNew.value = true
  testResult.value = null
}

function startEdit(r: FlagRule) {
  editing.value = { ...r }
  isNew.value = false
  testResult.value = null
  runTest()
}

function cancel() {
  editing.value = null
  isNew.value = false
  testResult.value = null
}

async function save() {
  if (!editing.value) return
  const r = editing.value
  try {
    if (isNew.value) {
      const created = await rulesApi.create({
        name: r.name,
        pattern: r.pattern,
        is_regex: r.is_regex,
        match_in: r.match_in,
        case_sensitive: r.case_sensitive,
        enabled: r.enabled,
      })
      rules.value = [created, ...rules.value]
      ui.showToast('Rule created', 'success')
    } else {
      const updated = await rulesApi.update(r.id, r)
      rules.value = rules.value.map((x) => (x.id === r.id ? updated : x))
      ui.showToast('Rule saved', 'success')
    }
    cancel()
  } catch (e) {
    ui.showToast('Failed to save rule', 'error')
  }
}

async function remove(r: FlagRule) {
  if (!confirm(`Delete rule "${r.name}"?`)) return
  try {
    await rulesApi.remove(r.id)
    rules.value = rules.value.filter((x) => x.id !== r.id)
    ui.showToast('Rule deleted', 'success')
  } catch {
    ui.showToast('Failed to delete', 'error')
  }
}

async function toggleEnabled(r: FlagRule) {
  try {
    const updated = await rulesApi.update(r.id, { enabled: !r.enabled })
    rules.value = rules.value.map((x) => (x.id === r.id ? updated : x))
  } catch {
    ui.showToast('Failed to toggle', 'error')
  }
}

let testTimer: number | null = null
function scheduleTest() {
  if (testTimer) clearTimeout(testTimer)
  testTimer = window.setTimeout(runTest, 350)
}

async function runTest() {
  if (!editing.value || !editing.value.pattern) {
    testResult.value = null
    return
  }
  testing.value = true
  try {
    testResult.value = await rulesApi.test({
      pattern: editing.value.pattern,
      is_regex: editing.value.is_regex,
      match_in: editing.value.match_in,
      case_sensitive: editing.value.case_sensitive,
    })
  } catch {
    testResult.value = null
  } finally {
    testing.value = false
  }
}

watch(
  () => editing.value && [editing.value.pattern, editing.value.is_regex, editing.value.match_in, editing.value.case_sensitive],
  () => scheduleTest(),
)

const rel = (s?: string | null) => (s ? dayjs.utc(s).fromNow() : '—')
</script>

<template>
  <div class="bg-noise min-h-full bg-ink-950 p-4 lg:p-6">
    <div class="mx-auto flex max-w-5xl flex-col gap-4">
      <header class="flex flex-wrap items-end justify-between gap-2">
        <div>
          <h1 class="display text-2xl font-semibold text-thaana-text">Rules</h1>
          <p class="mt-1 text-sm text-mid">Patterns that flag items automatically.</p>
        </div>
        <button class="btn-primary" @click="startNew">
          <Plus :size="14" :stroke-width="2" />
          Add rule
        </button>
      </header>

      <section v-if="editing" class="card flex flex-col gap-4 p-4 lg:p-5">
        <header class="flex items-center justify-between">
          <h2 class="display text-sm font-semibold text-thaana-text">
            {{ isNew ? 'New rule' : 'Edit rule' }}
          </h2>
          <button class="btn-icon" @click="cancel"><X :size="14" :stroke-width="1.5" /></button>
        </header>

        <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
          <div class="flex flex-col gap-1.5">
            <label class="text-[10px] font-medium uppercase tracking-wider text-dim">Name</label>
            <input v-model="editing.name" type="text" class="input-base" placeholder="e.g. Tender deadline" />
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="text-[10px] font-medium uppercase tracking-wider text-dim">Match in</label>
            <select v-model="editing.match_in" class="input-base">
              <option value="all">All fields</option>
              <option value="title">Title</option>
              <option value="body">Body</option>
              <option value="url">URL</option>
            </select>
          </div>
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-[10px] font-medium uppercase tracking-wider text-dim">Pattern</label>
          <input
            v-model="editing.pattern"
            type="text"
            class="input-base font-mono text-xs"
            placeholder="e.g. /tender|vacancy/i or just plain text"
          />
        </div>

        <div class="flex flex-wrap items-center gap-3 text-xs text-mid">
          <label class="flex cursor-pointer items-center gap-2">
            <input
              type="checkbox"
              v-model="editing.is_regex"
              class="h-3.5 w-3.5 rounded-sm border-ink-700 bg-transparent accent-violet-400"
            />
            Treat as regex
          </label>
          <label class="flex cursor-pointer items-center gap-2">
            <input
              type="checkbox"
              v-model="editing.case_sensitive"
              class="h-3.5 w-3.5 rounded-sm border-ink-700 bg-transparent accent-violet-400"
            />
            Case sensitive
          </label>
          <label class="flex cursor-pointer items-center gap-2">
            <input
              type="checkbox"
              v-model="editing.enabled"
              class="h-3.5 w-3.5 rounded-sm border-ink-700 bg-transparent accent-violet-400"
            />
            Enabled
          </label>
        </div>

        <div class="rounded-lg border border-ink-700 bg-ink-800/40 p-3">
          <div class="mb-2 flex items-center gap-2 text-[10px] uppercase tracking-wider text-dim">
            <FlaskConical :size="10" :stroke-width="1.5" />
            Live test
          </div>
          <div v-if="!editing.pattern" class="text-xs text-dim">Type a pattern to see matches.</div>
          <div v-else-if="testing" class="text-xs text-mid">Testing…</div>
          <div v-else-if="testResult" class="flex flex-col gap-2">
            <div class="text-sm">
              <span class="num text-thaana-text">{{ testResult.matches.toLocaleString() }}</span>
              <span class="text-mid"> of </span>
              <span class="num text-thaana-text">{{ testResult.total.toLocaleString() }}</span>
              <span class="text-mid"> items match</span>
            </div>
            <ul v-if="testResult.sample.length" class="flex flex-col gap-1.5">
              <li v-for="s in testResult.sample" :key="s.id" class="rounded-md border border-ink-700 bg-ink-900 p-2 text-xs">
                <div class="truncate text-thaana-text">{{ s.title }}</div>
                <div v-if="s.snippet" class="mt-0.5 truncate text-mid">{{ s.snippet }}</div>
              </li>
            </ul>
          </div>
        </div>

        <footer class="flex items-center justify-end gap-2 border-t border-line pt-3">
          <button class="btn-ghost" @click="cancel">Cancel</button>
          <button class="btn-primary" :disabled="!editing.name || !editing.pattern" @click="save">
            <Save :size="14" :stroke-width="1.75" />
            {{ isNew ? 'Create' : 'Save' }}
          </button>
        </footer>
      </section>

      <div v-if="loading" class="flex flex-col gap-2">
        <Skeleton v-for="i in 4" :key="i" height="80px" />
      </div>

      <EmptyState
        v-else-if="!rules.length && !editing"
        variant="flags"
        title="No rules yet"
        subtitle="Add your first pattern to start flagging items."
      />

      <ul v-else class="flex flex-col gap-2">
        <li
          v-for="r in rules"
          :key="r.id"
          class="card flex flex-col gap-2 p-3 lg:flex-row lg:items-center lg:gap-3 lg:p-4"
        >
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="display truncate text-sm font-semibold text-thaana-text">{{ r.name }}</span>
              <span
                v-if="r.enabled"
                class="inline-flex h-5 items-center rounded-md bg-signal-green/15 px-1.5 text-[10px] font-medium text-signal-green"
              >
                on
              </span>
              <span
                v-else
                class="inline-flex h-5 items-center rounded-md bg-ink-700 px-1.5 text-[10px] font-medium text-dim"
              >
                off
              </span>
            </div>
            <code class="mt-1 block truncate font-mono text-xs text-mid">{{ r.pattern }}</code>
            <div class="mt-1 flex flex-wrap gap-2 text-[10px] text-dim">
              <span>match: {{ r.match_in }}</span>
              <span v-if="r.is_regex">regex</span>
              <span v-if="r.case_sensitive">case-sensitive</span>
              <span class="num">{{ r.last_match_count || 0 }} matches</span>
              <span v-if="r.last_match_at" class="num">last {{ rel(r.last_match_at) }}</span>
            </div>
          </div>
          <div class="flex items-center gap-1.5 self-end lg:self-auto">
            <button class="btn-icon" :title="r.enabled ? 'Disable' : 'Enable'" @click="toggleEnabled(r)">
              <component :is="r.enabled ? ToggleRight : ToggleLeft" :size="16" :stroke-width="1.5" />
            </button>
            <button class="btn-ghost" @click="startEdit(r)">Edit</button>
            <button class="btn-icon text-signal-rose" title="Delete" @click="remove(r)">
              <Trash2 :size="14" :stroke-width="1.5" />
            </button>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>
