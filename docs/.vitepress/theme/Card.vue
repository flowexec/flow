<script setup lang="ts">
import Icon from './Icon.vue'

defineProps<{
  icon: string
  title: string
  badge?: string
  stack?: string[]
  href?: string
  hrefText?: string
  altHref?: string
  altHrefText?: string
}>()
</script>

<template>
  <article class="card">
    <div class="card__head">
      <span class="card__icon"><Icon :name="icon" /></span>
      <div class="card__heading">
        <h3 class="card__title">{{ title }}</h3>
        <span v-if="badge" class="card__badge">{{ badge }}</span>
      </div>
    </div>

    <p class="card__body"><slot /></p>

    <p v-if="stack?.length" class="card__stack">
      <template v-for="(item, i) in stack" :key="item">
        <span v-if="i > 0" class="sep">·</span>{{ item }}
      </template>
    </p>

    <p v-if="href" class="card__links">
      <a :href="href">{{ hrefText ?? 'Guide' }}</a>
      <a v-if="altHref" :href="altHref">{{ altHrefText ?? 'Reference' }}</a>
    </p>
  </article>
</template>

<style scoped>
.card {
  display: flex;
  flex-direction: column;
  padding: 22px;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-border);
  border-radius: var(--radius);
  transition: border-color 0.18s ease;
}

.card:hover {
  border-color: var(--vp-c-text-3);
}

.card__head {
  display: flex;
  align-items: center;
  gap: 11px;
  margin-bottom: 10px;
}

.card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  flex: 0 0 auto;
  border: 1px solid var(--vp-c-border);
  border-radius: 8px;
  font-size: 15px;
  color: var(--vp-c-brand-1);
}

.card__heading {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  min-width: 0;
}

.card__title {
  margin: 0;
  padding: 0;
  border: 0;
  font-size: 17px;
  font-weight: 600;
  line-height: 1.3;
  letter-spacing: -0.01em;
  color: var(--vp-c-text-1);
}

.card__badge {
  padding: 2px 8px;
  border-radius: 6px;
  background: var(--vp-c-default-soft);
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--vp-c-text-2);
}

.card__body {
  flex: 1 1 auto;
  margin: 0 0 14px;
  font-size: 14px;
  line-height: 1.62;
  color: var(--vp-c-text-2);
}

.card__stack {
  margin: 0 0 14px;
  font-size: 12.5px;
  letter-spacing: 0.02em;
  color: var(--vp-c-text-3);
}

.card__stack .sep {
  margin: 0 7px;
  opacity: 0.5;
}

.card__links {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin: 0;
  font-size: 13.5px;
}

.card__links a {
  color: var(--vp-c-brand-1);
  font-weight: 500;
  text-decoration: none;
  box-shadow: 0 1px 0 var(--vp-c-brand-1);
  transition: opacity 0.18s ease;
}

.card__links a:hover {
  opacity: 0.75;
}
</style>
