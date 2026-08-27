<script setup lang="ts">
import Icon from './Icon.vue'

defineProps<{
  icon: string
  label: string
  moreText?: string
  moreLink?: string
}>()
</script>

<template>
  <h2 class="home-section__title">
    <span class="home-section__label">
      <Icon :name="icon" />
      {{ label }}
    </span>
    <a v-if="moreLink" class="home-section__more" :href="moreLink">
      {{ moreText ?? 'More' }}
      <Icon name="arrowRight" />
    </a>
  </h2>
</template>

<style scoped>
h2.home-section__title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: calc(var(--gap) * 1.6) 0 calc(var(--gap) * 0.8);
  padding-bottom: 10px;
  border: 0;
  border-bottom: 1px solid var(--vp-c-border);
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--vp-c-text-2);
}

.home-section__label {
  display: inline-flex;
  align-items: center;
  gap: 9px;
}

.home-section__label :deep(.ui-icon) {
  width: 1.05em;
  height: 1.05em;
}

.home-section__more {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.04em;
  text-transform: none;
  white-space: nowrap;
  color: var(--vp-c-text-2);
  text-decoration: none;
  transition: color 0.18s ease;
}

.home-section__more:hover {
  color: var(--vp-c-brand-1);
}

@media (max-width: 600px) {
  h2.home-section__title {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
}
</style>
