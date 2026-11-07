<script setup lang="ts">
import { computed } from 'vue'
import Icon from './Icon.vue'

// The counterpart of the FlowRef callout on Mochi's docs: every hand-off to
// Mochi renders through this one component, so the sibling link looks the same
// in both directions and each placement is attributable.
const props = withDefaults(defineProps<{
  /** Where on flowexec.io the callout sits; sent as utm_content. */
  placement: string
  title: string
  /** Path on mochiexec.io, e.g. `/`. */
  path?: string
  eyebrow?: string
}>(), {
  path: '/',
  eyebrow: 'Mochi',
})

const href = computed(() => {
  const url = new URL(props.path, 'https://mochiexec.io')
  url.searchParams.set('utm_source', 'flowexec.io')
  url.searchParams.set('utm_medium', 'docs')
  url.searchParams.set('utm_content', props.placement)
  return url.toString()
})
</script>

<template>
  <aside class="mochi-ref">
    <div class="mochi-ref__head">
      <span class="mochi-ref__eyebrow">{{ eyebrow }}</span>
      <a class="mochi-ref__link" :href="href">
        {{ title }}
        <Icon name="external" />
      </a>
    </div>
    <div class="mochi-ref__body"><slot /></div>
  </aside>
</template>

<style scoped>
.mochi-ref {
  margin: 20px 0;
  padding: 14px 18px;
  border: 1px solid var(--vp-c-border);
  border-left: 3px solid var(--mochi-accent);
  border-radius: var(--radius);
  background: var(--vp-c-bg-soft);
}

.mochi-ref__head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 4px 12px;
}

.mochi-ref__eyebrow {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--mochi-accent);
}

.mochi-ref__link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-weight: 600;
  color: var(--vp-c-text-1);
  text-decoration: none;
}

.mochi-ref__link:hover {
  color: var(--mochi-accent);
}

.mochi-ref__body {
  font-size: 14px;
  line-height: 1.62;
  color: var(--vp-c-text-2);
}

.mochi-ref__body :deep(p) {
  margin: 6px 0 0;
}
</style>
