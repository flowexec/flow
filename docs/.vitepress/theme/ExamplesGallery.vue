<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

const copied = ref(false)
async function copySource(text: string) {
  await navigator.clipboard.writeText(text)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}
import { data } from '../../examples.data.js'
import type { FlowFile } from '../../examples.data.js'

const CATEGORY_LABELS: Record<string, string> = {
  basics: 'Basics',
  general: 'Basics',
  'go-project': 'Go Project',
  docker: 'Docker',
  setup: 'Setup',
  git: 'Git',
  api: 'API',
  kubernetes: 'Kubernetes',
}

function categoryLabel(cat: string) {
  return CATEGORY_LABELS[cat] ?? cat.charAt(0).toUpperCase() + cat.slice(1)
}

const activeCategory = ref('all')
const modalFile = ref<FlowFile | null>(null)

const filteredFiles = computed(() =>
  activeCategory.value === 'all'
    ? data.files
    : data.files.filter((f) => f.category === activeCategory.value)
)

function openModal(file: FlowFile) {
  modalFile.value = file
  document.body.style.overflow = 'hidden'
}

function closeModal() {
  modalFile.value = null
  document.body.style.overflow = ''
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') closeModal()
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
})
</script>

<template>
  <div class="examples-gallery">
    <div class="category-filters">
      <button
        :class="['filter-btn', { active: activeCategory === 'all' }]"
        @click="activeCategory = 'all'"
      >
        All
      </button>
      <button
        v-for="cat in data.categories"
        :key="cat"
        :class="['filter-btn', { active: activeCategory === cat }]"
        @click="activeCategory = cat"
      >
        {{ categoryLabel(cat) }}
      </button>
    </div>

    <div class="examples-grid">
      <div v-for="file in filteredFiles" :key="file.path" class="example-card">
        <div class="card-header">
          <span class="category-badge">{{ categoryLabel(file.category) }}</span>
          <a
            :href="file.sourceUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="github-link"
            title="View on GitHub"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="15"
              height="15"
              viewBox="0 0 24 24"
              fill="currentColor"
              aria-hidden="true"
            >
              <path
                d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"
              />
            </svg>
          </a>
        </div>

        <h3 class="card-title">{{ file.namespace }}</h3>
        <p class="card-description">{{ file.description }}</p>

        <div v-if="file.tags.length" class="tags">
          <span v-for="tag in file.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>

        <div v-if="file.executables.length" class="executables">
          <span
            v-for="ex in file.executables"
            :key="`${ex.verb}:${ex.name ?? ''}`"
            :title="ex.description"
            class="exec-badge"
          >
            {{ ex.verb }}<template v-if="ex.name"> {{ ex.name }}</template>
          </span>
        </div>

        <div class="card-footer">
          <button class="source-btn" @click="openModal(file)">View source</button>
        </div>
      </div>
    </div>
  </div>

  <!-- Modal -->
  <Teleport to="body">
    <div v-if="modalFile" class="modal-backdrop" @click.self="closeModal">
      <div class="modal" role="dialog" aria-modal="true">
        <div class="modal-header">
          <div class="modal-title-group">
            <span class="modal-namespace">{{ modalFile.namespace }}</span>
            <span class="modal-category">{{ categoryLabel(modalFile.category) }}</span>
          </div>
          <div class="modal-actions">
            <button class="modal-copy-btn" @click="copySource(modalFile.rawContent)">
              {{ copied ? '✓ Copied' : 'Copy' }}
            </button>
            <a
              :href="modalFile.sourceUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="modal-github-link"
              title="View on GitHub"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="currentColor"
                aria-hidden="true"
              >
                <path
                  d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"
                />
              </svg>
              GitHub
            </a>
            <button class="modal-close" @click="closeModal" aria-label="Close">✕</button>
          </div>
        </div>
        <div class="modal-body" v-html="modalFile.highlightedContent" />
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.examples-gallery {
  margin-top: 1.5rem;
}

.category-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1.75rem;
}

.filter-btn {
  padding: 0.3rem 0.9rem;
  border-radius: 20px;
  border: 1px solid var(--vp-c-border);
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-text-2);
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.15s;
  line-height: 1.5;
}

.filter-btn:hover {
  border-color: var(--vp-c-brand-1);
  color: var(--vp-c-brand-1);
}

.filter-btn.active {
  background: var(--vp-c-brand-1);
  border-color: var(--vp-c-brand-1);
  color: #fff;
}

.examples-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

.example-card {
  border: 1px solid var(--vp-c-border);
  border-radius: 10px;
  padding: 1rem 1.1rem;
  background: var(--vp-c-bg-soft);
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.category-badge {
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.03em;
  text-transform: uppercase;
  padding: 0.15rem 0.6rem;
  border-radius: 20px;
  background: var(--vp-c-brand-1);
  color: #fff;
}

.github-link {
  color: var(--vp-c-text-3);
  display: flex;
  align-items: center;
  transition: color 0.15s;
}

.github-link:hover {
  color: var(--vp-c-brand-1);
}

.card-title {
  margin: 0 !important;
  padding: 0 !important;
  font-size: 1rem !important;
  font-weight: 600;
  border: none !important;
  font-family: var(--vp-font-family-mono);
  color: var(--vp-c-text-1);
  line-height: 1.4;
}

.card-description {
  margin: 0;
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  line-height: 1.55;
  flex: 1;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
}

.tag {
  font-size: 0.72rem;
  padding: 0.1rem 0.45rem;
  border-radius: 4px;
  background: var(--vp-c-bg-mute);
  color: var(--vp-c-text-2);
  border: 1px solid var(--vp-c-border);
}

.executables {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
}

.exec-badge {
  font-size: 0.74rem;
  font-family: var(--vp-font-family-mono);
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  background: var(--vp-c-brand-2);
  color: #fff;
  cursor: default;
}

.card-footer {
  padding-top: 0.5rem;
  border-top: 1px solid var(--vp-c-divider);
}

.source-btn {
  font-size: 0.8rem;
  color: var(--vp-c-brand-1);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  transition: opacity 0.15s;
}

.source-btn:hover {
  opacity: 0.7;
}

/* Modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal {
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-border);
  border-radius: 12px;
  width: 100%;
  max-width: 780px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.9rem 1.2rem;
  border-bottom: 1px solid var(--vp-c-divider);
  flex-shrink: 0;
  gap: 1rem;
}

.modal-title-group {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  min-width: 0;
}

.modal-namespace {
  font-family: var(--vp-font-family-mono);
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--vp-c-text-1);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.modal-category {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.03em;
  text-transform: uppercase;
  padding: 0.15rem 0.55rem;
  border-radius: 20px;
  background: var(--vp-c-brand-1);
  color: #fff;
  flex-shrink: 0;
}

.modal-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-shrink: 0;
}

.modal-copy-btn {
  font-size: 0.8rem;
  padding: 0.25rem 0.7rem;
  border-radius: 6px;
  border: 1px solid var(--vp-c-border);
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-text-2);
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.modal-copy-btn:hover {
  border-color: var(--vp-c-brand-1);
  color: var(--vp-c-brand-1);
}

.modal-github-link {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.82rem;
  color: var(--vp-c-text-2);
  text-decoration: none;
  transition: color 0.15s;
}

.modal-github-link:hover {
  color: var(--vp-c-brand-1);
}

.modal-close {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1rem;
  color: var(--vp-c-text-3);
  padding: 0.15rem 0.3rem;
  border-radius: 4px;
  line-height: 1;
  transition: color 0.15s;
}

.modal-close:hover {
  color: var(--vp-c-text-1);
}

.modal-body {
  overflow-y: auto;
  flex: 1;
  padding: 1rem;
}

.modal-body :deep(div[class*='language-']) {
  margin: 0 !important;
  border-radius: 8px !important;
}

.modal-body :deep(div[class*='language-'] pre) {
  max-height: none !important;
  padding: 1.25rem 1.5rem !important;
}
</style>
