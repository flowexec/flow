<script setup lang="ts">
import { ref } from 'vue'
import Icon from './Icon.vue'
import { data as release } from '../../release.data.js'

const INSTALL_CMD = 'curl -sSL https://install.flowexec.io | bash'

const copied = ref(false)
async function copyInstall() {
  await navigator.clipboard.writeText(INSTALL_CMD)
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 2000)
}
</script>

<template>
  <ul class="now-strip">
    <li class="now-item">
      <Icon name="terminal" />
      <div class="now-item__body">
        <span class="now-item__label">Install</span>
        <button
          class="now-item__cmd"
          :title="copied ? 'Copied' : 'Copy to clipboard'"
          @click="copyInstall"
        >
          <code>{{ INSTALL_CMD }}</code>
          <Icon :name="copied ? 'check' : 'copy'" />
        </button>
      </div>
    </li>

    <li class="now-item">
      <Icon name="layers" />
      <div class="now-item__body">
        <span class="now-item__label">Latest</span>
        <span class="now-item__value tabular">
          <a :href="release.url">{{ release.tag || 'Releases' }}</a>
        </span>
      </div>
    </li>

    <li class="now-item">
      <Icon name="monitor" />
      <div class="now-item__body">
        <span class="now-item__label">Runs on</span>
        <span class="now-item__value">macOS · Linux · Windows</span>
      </div>
    </li>
  </ul>
</template>

<style scoped>
/* The 2px gap over a border-coloured ground *is* the hairline divider — one
   rule instead of per-cell borders that double up at the seams. */
.now-strip {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) repeat(auto-fit, minmax(150px, 1fr));
  gap: 1px;
  margin: 0 0 calc(var(--gap) * 1.2);
  padding: 0;
  list-style: none;
  background: var(--vp-c-border);
  border: 1px solid var(--vp-c-border);
  border-radius: var(--radius);
  overflow: hidden;
}

.now-item {
  display: flex;
  align-items: flex-start;
  gap: 11px;
  margin: 0;
  padding: 15px 17px;
  background: var(--vp-c-bg);
  font-size: 14px;
}

.now-item > :deep(.ui-icon) {
  width: 15px;
  height: 15px;
  margin-top: 3px;
  color: var(--vp-c-text-2);
}

.now-item__body {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.now-item__label {
  font-size: 10.5px;
  line-height: 1.4;
  letter-spacing: 0.09em;
  text-transform: uppercase;
  color: var(--vp-c-text-2);
}

.now-item__value {
  line-height: 1.45;
  color: var(--vp-c-text-1);
}

.now-item__value a {
  color: var(--vp-c-brand-1);
  text-decoration: none;
  box-shadow: 0 1px 0 var(--vp-c-brand-1);
}

.now-item__cmd {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  min-width: 0;
  padding: 0;
  border: 0;
  background: none;
  color: var(--vp-c-text-1);
  cursor: pointer;
  text-align: left;
}

.now-item__cmd code {
  overflow: hidden;
  font-family: var(--vp-font-family-mono);
  font-size: 12.5px;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.now-item__cmd :deep(.ui-icon) {
  width: 14px;
  height: 14px;
  color: var(--vp-c-text-3);
  transition: color 0.18s ease;
}

.now-item__cmd:hover :deep(.ui-icon) {
  color: var(--vp-c-brand-1);
  opacity: 1;
}

@media (max-width: 700px) {
  .now-strip {
    grid-template-columns: 1fr;
  }
}
</style>
