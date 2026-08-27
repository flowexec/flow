<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref } from 'vue'
import Icon from './Icon.vue'
import { openAsk } from './askState'

/**
 * A row inside VitePress's own search modal that hands the current query over
 * to the Ask panel — the "no page matched, ask the codebase" path.
 *
 * VitePress teleports its search box to <body> and exposes no slot, so the
 * only way in is to observe the DOM. That means depending on three internal
 * selectors, so every one of them is checked before the row renders: if a
 * VitePress upgrade renames any of them the bridge quietly disappears and the
 * standalone Ask button carries on working.
 */
const SELECTORS = {
  box: '.VPLocalSearchBox',
  shell: '.VPLocalSearchBox .shell',
  input: '#localsearch-input',
  backdrop: '.VPLocalSearchBox .backdrop',
}

const ready = ref(false)
const query = ref('')

let observer: MutationObserver | null = null
let input: HTMLInputElement | null = null

function syncQuery() {
  query.value = input?.value.trim() ?? ''
}

function attach() {
  if (ready.value) return
  const shell = document.querySelector(SELECTORS.shell)
  const found = document.querySelector<HTMLInputElement>(SELECTORS.input)
  if (!shell || !found) return

  input = found
  input.addEventListener('input', syncQuery)
  syncQuery()
  ready.value = true
}

function detach() {
  input?.removeEventListener('input', syncQuery)
  input = null
  ready.value = false
  query.value = ''
}

function onMutation() {
  if (document.querySelector(SELECTORS.box)) attach()
  else if (ready.value) detach()
}

async function handoff() {
  const q = query.value
  // Drop the teleport before VitePress removes the node it renders into.
  ready.value = false
  await nextTick()
  document.querySelector<HTMLElement>(SELECTORS.backdrop)?.click()
  openAsk(q)
}

onMounted(() => {
  observer = new MutationObserver(onMutation)
  observer.observe(document.body, { childList: true, subtree: true })
})

onUnmounted(() => {
  observer?.disconnect()
  observer = null
  detach()
})
</script>

<template>
  <Teleport v-if="ready" :to="SELECTORS.shell">
    <button class="ask-bridge" type="button" @click="handoff">
      <Icon name="sparkle" />
      <span class="ask-bridge__text">
        <template v-if="query">Ask the codebase about “{{ query }}”</template>
        <template v-else>Ask the codebase a question</template>
      </span>
      <Icon name="arrowRight" />
    </button>
  </Teleport>
</template>

<style scoped>
.ask-bridge {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  border: 0;
  border-top: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-text-2);
  font-size: 13.5px;
  text-align: left;
  cursor: pointer;
  transition: color 0.18s ease, background-color 0.18s ease;
}

.ask-bridge:hover {
  background: var(--vp-c-default-soft);
  color: var(--vp-c-text-1);
}

.ask-bridge__text {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ask-bridge :deep(.ui-icon) {
  width: 14px;
  height: 14px;
}

.ask-bridge :deep(.ui-icon:first-child) {
  color: var(--vp-c-brand-1);
  opacity: 1;
}
</style>
