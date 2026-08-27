<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import MarkdownIt from 'markdown-it'
import Icon from './Icon.vue'
import { askDeepWiki, DeepWikiError, DEEPWIKI_URL, REPO } from './deepwiki'
import { askOpen, askQuery } from './askState'

const SUGGESTIONS = [
  'How does flow resolve an executable reference?',
  'What happens when a parallel executable fails?',
  'How are secrets injected into a running executable?',
]

// `html: false` is the sanitisation boundary — model output never reaches the
// DOM as markup, only as text markdown-it itself escaped.
const md = new MarkdownIt({ html: false, linkify: true, breaks: false })

const input = ref<HTMLInputElement | null>(null)
const question = ref('')
const answer = ref('')
const error = ref('')
const loading = ref(false)
const elapsed = ref(0)
const asked = ref('')

let controller: AbortController | null = null

const rendered = computed(() => (answer.value ? md.render(answer.value) : ''))

function cancel() {
  controller?.abort()
  controller = null
  loading.value = false
}

function close() {
  cancel()
  askOpen.value = false
}

function reset() {
  answer.value = ''
  error.value = ''
  elapsed.value = 0
  asked.value = ''
}

async function submit() {
  const q = question.value.trim()
  if (!q || loading.value) return

  reset()
  loading.value = true
  asked.value = q
  controller = new AbortController()

  try {
    answer.value = await askDeepWiki(q, {
      signal: controller.signal,
      onProgress: (seconds) => {
        elapsed.value = seconds
      },
    })
  } catch (err) {
    if (controller?.signal.aborted) return
    error.value =
      err instanceof DeepWikiError
        ? err.message
        : 'Could not reach DeepWiki. It may be rate limited or temporarily down.'
  } finally {
    loading.value = false
    controller = null
  }
}

function ask(suggestion: string) {
  question.value = suggestion
  submit()
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && askOpen.value) {
    e.preventDefault()
    close()
  }
}

// Layout mounts this component only while the panel is open, so the component
// lifetime *is* the open state — a watcher on askOpen would register too late
// to ever see the transition that created it.
onMounted(async () => {
  document.addEventListener('keydown', onKeydown)
  document.body.style.overflow = 'hidden'

  // A query handed over from the search modal arrives here.
  if (askQuery.value) {
    question.value = askQuery.value
    askQuery.value = ''
  }

  await nextTick()
  input.value?.focus()
  input.value?.select()
})
onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
  controller?.abort()
})
</script>

<template>
  <Teleport to="body">
    <div v-if="askOpen" class="ask" role="dialog" aria-modal="true" aria-label="Ask the flow codebase">
      <div class="ask__backdrop" @click="close" />

      <div class="ask__shell">
        <form class="ask__bar" @submit.prevent="submit">
          <Icon name="sparkle" />
          <input
            ref="input"
            v-model="question"
            class="ask__input"
            type="search"
            placeholder="Ask anything about the flow codebase…"
            autocomplete="off"
            spellcheck="false"
          />
          <button
            v-if="!loading"
            class="ask__btn"
            type="submit"
            title="Ask"
            :disabled="!question.trim()"
          >
            <Icon name="enter" />
          </button>
          <button
            v-else
            class="ask__btn"
            type="button"
            title="Stop"
            @click="cancel"
          >
            <span class="ask__stop" />
          </button>

          <button class="ask__btn ask__close" type="button" title="Close" @click="close">
            <Icon name="close" />
          </button>
        </form>

        <div class="ask__body">
          <!-- Idle -->
          <div v-if="!loading && !answer && !error" class="ask__idle">
            <p class="ask__idle-lede">
              Answers are grounded in the <code>{{ REPO }}</code> source, not just these docs.
              Expect about 15 seconds.
            </p>
            <ul class="ask__suggestions">
              <li v-for="s in SUGGESTIONS" :key="s">
                <button type="button" @click="ask(s)">
                  {{ s }}
                  <Icon name="arrowRight" />
                </button>
              </li>
            </ul>
          </div>

          <!-- Working -->
          <div v-else-if="loading" class="ask__loading">
            <p class="ask__status">
              <span class="ask__pulse" />
              Reading the flow codebase<span v-if="elapsed" class="tabular"> · {{ elapsed }}s</span>
            </p>
            <div class="ask__skeleton">
              <span v-for="n in 5" :key="n" :style="{ width: `${[96, 88, 72, 91, 54][n - 1]}%` }" />
            </div>
          </div>

          <!-- Failed -->
          <div v-else-if="error" class="ask__error">
            <p><Icon name="alert" /> {{ error }}</p>
            <a :href="DEEPWIKI_URL" target="_blank" rel="noopener noreferrer">
              Ask on DeepWiki instead <Icon name="external" />
            </a>
          </div>

          <!-- Answered -->
          <div v-else class="ask__answer">
            <p class="ask__asked">{{ asked }}</p>
            <div class="vp-doc" v-html="rendered" />
          </div>
        </div>

        <div class="ask__foot">
          <span>
            Generated by
            <a :href="DEEPWIKI_URL" target="_blank" rel="noopener noreferrer">DeepWiki</a>
            from the flow source · may be inaccurate
          </span>
          <span class="ask__foot-keys"><kbd>esc</kbd> close</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.ask {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  justify-content: center;
  padding: 64px 16px 16px;
}

.ask__backdrop {
  position: absolute;
  inset: 0;
  background: var(--vp-backdrop-bg-color);
  backdrop-filter: blur(4px);
}

.ask__shell {
  position: relative;
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 720px;
  max-height: 100%;
  overflow: hidden;
  background: var(--vp-c-bg-elv);
  border: 1px solid var(--vp-c-border);
  border-radius: var(--radius);
}

.ask__bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--vp-c-divider);
}

.ask__bar > :deep(.ui-icon) {
  width: 17px;
  height: 17px;
  color: var(--vp-c-brand-1);
  opacity: 1;
}

.ask__input {
  flex: 1 1 auto;
  min-width: 0;
  border: 0;
  background: transparent;
  color: var(--vp-c-text-1);
  font-size: 16px;
  outline: none;
}

.ask__input::placeholder {
  color: var(--vp-c-text-3);
}

.ask__input::-webkit-search-cancel-button {
  display: none;
}

.ask__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  flex: 0 0 auto;
  border: 1px solid var(--vp-c-border);
  border-radius: 7px;
  background: transparent;
  color: var(--vp-c-text-2);
  cursor: pointer;
  transition: border-color 0.18s ease, color 0.18s ease;
}

.ask__btn:hover:not(:disabled) {
  border-color: var(--vp-c-text-3);
  color: var(--vp-c-text-1);
}

.ask__btn:disabled {
  opacity: 0.4;
  cursor: default;
}

.ask__btn :deep(.ui-icon) {
  width: 14px;
  height: 14px;
}

.ask__close {
  border-color: transparent;
}

.ask__stop {
  width: 9px;
  height: 9px;
  border-radius: 2px;
  background: currentColor;
}

.ask__body {
  flex: 1 1 auto;
  overflow-y: auto;
  padding: 18px 20px;
}

/* --- idle ---------------------------------------------------------------- */

.ask__idle-lede {
  margin: 0 0 16px;
  font-size: 13.5px;
  line-height: 1.62;
  color: var(--vp-c-text-2);
}

.ask__idle-lede code {
  font-family: var(--vp-font-family-mono);
  font-size: 0.9em;
}

.ask__suggestions {
  list-style: none;
  margin: 0;
  padding: 0;
}

.ask__suggestions li + li {
  border-top: 1px solid var(--vp-c-divider);
}

.ask__suggestions button {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  padding: 11px 2px;
  border: 0;
  background: none;
  color: var(--vp-c-text-1);
  font-size: 14px;
  text-align: left;
  cursor: pointer;
  transition: color 0.18s ease;
}

.ask__suggestions button:hover {
  color: var(--vp-c-brand-1);
}

.ask__suggestions :deep(.ui-icon) {
  width: 14px;
  height: 14px;
}

/* --- working ------------------------------------------------------------- */

.ask__status {
  display: flex;
  align-items: center;
  gap: 9px;
  margin: 0 0 16px;
  font-size: 12.5px;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--vp-c-text-2);
}

.ask__pulse {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--vp-c-brand-1);
  animation: ask-pulse 1.4s ease-in-out infinite;
}

@keyframes ask-pulse {
  0%,
  100% {
    opacity: 0.25;
    transform: scale(0.8);
  }
  50% {
    opacity: 1;
    transform: scale(1);
  }
}

.ask__skeleton {
  display: flex;
  flex-direction: column;
  gap: 11px;
}

.ask__skeleton span {
  height: 9px;
  border-radius: 5px;
  background: var(--vp-c-default-soft);
  animation: ask-shimmer 1.6s ease-in-out infinite;
}

.ask__skeleton span:nth-child(2) { animation-delay: 0.1s; }
.ask__skeleton span:nth-child(3) { animation-delay: 0.2s; }
.ask__skeleton span:nth-child(4) { animation-delay: 0.3s; }
.ask__skeleton span:nth-child(5) { animation-delay: 0.4s; }

@keyframes ask-shimmer {
  0%,
  100% { opacity: 0.45; }
  50% { opacity: 1; }
}

/* --- failed -------------------------------------------------------------- */

.ask__error p {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 12px;
  font-size: 14px;
  color: var(--vp-c-text-1);
}

.ask__error a {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13.5px;
  font-weight: 500;
  color: var(--vp-c-brand-1);
  text-decoration: none;
}

.ask__error :deep(.ui-icon) {
  width: 15px;
  height: 15px;
}

/* --- answered ------------------------------------------------------------ */

.ask__asked {
  margin: 0 0 14px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--vp-c-divider);
  font-size: 13px;
  letter-spacing: 0.02em;
  color: var(--vp-c-text-2);
}

.ask__answer :deep(.vp-doc) {
  font-size: 14.5px;
}

.ask__answer :deep(.vp-doc > *:first-child) {
  margin-top: 0;
}

.ask__answer :deep(pre) {
  padding: 14px 16px;
  overflow-x: auto;
  background: var(--vp-code-block-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  font-family: var(--vp-font-family-mono);
  font-size: 12.5px;
  line-height: 1.6;
}

.ask__answer :deep(h2),
.ask__answer :deep(h3) {
  margin: 1.6em 0 0.6em;
  padding: 0;
  border: 0;
  font-size: 15px;
  font-weight: 600;
}

/* --- foot ---------------------------------------------------------------- */

.ask__foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 16px;
  border-top: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg-soft);
  font-size: 11.5px;
  color: var(--vp-c-text-3);
}

.ask__foot a {
  color: var(--vp-c-text-2);
}

.ask__foot kbd {
  padding: 1px 5px;
  border: 1px solid var(--vp-c-border);
  border-radius: 4px;
  font-family: var(--vp-font-family-mono);
  font-size: 10px;
}

@media (max-width: 640px) {
  .ask {
    padding: 12px;
  }
  .ask__foot-keys {
    display: none;
  }
}
</style>
