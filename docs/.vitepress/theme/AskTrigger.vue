<script setup lang="ts">
import Icon from './Icon.vue'
import { openAsk } from './askState'
</script>

<template>
  <button class="ask-trigger" title="Ask about the flow codebase (⌘I)" @click="openAsk()">
    <Icon name="sparkle" />
    <span class="ask-trigger__text">Ask AI</span>
  </button>
</template>

<style scoped>
.ask-trigger {
  display: inline-flex;
  flex: 0 0 auto;
  margin-left: 8px;
  align-items: center;
  gap: 6px;
  padding: 0 10px;
  height: 30px;
  border: 1px solid var(--vp-c-border);
  border-radius: 8px;
  background: transparent;
  color: var(--vp-c-text-2);
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  cursor: pointer;
  transition: border-color 0.18s ease, color 0.18s ease;
}

.ask-trigger:hover {
  border-color: var(--vp-c-text-3);
  color: var(--vp-c-text-1);
}

.ask-trigger :deep(.ui-icon) {
  width: 14px;
  height: 14px;
  color: var(--vp-c-brand-1);
  opacity: 1;
}

/* In the mobile nav screen the trigger is a full-width row, not a pill. */
.VPNavScreen .ask-trigger {
  width: 100%;
  height: 38px;
  justify-content: center;
}
</style>
