<script setup lang="ts">
defineProps<{
  index: string
  label: string
  claim: string
  /** The "not another X — Y" line that heads off the obvious misreading. */
  aside?: string
}>()
</script>

<template>
  <section class="numbered">
    <div class="numbered__rail">
      <span class="numbered__index tabular">{{ index }}</span>
      <span class="numbered__label">{{ label }}</span>
    </div>
    <div class="numbered__body">
      <p class="numbered__claim">{{ claim }}</p>
      <p class="numbered__text"><slot /></p>
      <p v-if="aside" class="numbered__aside">{{ aside }}</p>
    </div>
  </section>
</template>

<style scoped>
.numbered {
  display: grid;
  grid-template-columns: 6.5rem minmax(0, 1fr);
  gap: 0 24px;
  padding: 26px 0;
  border-top: 1px solid var(--vp-c-divider);
}

.numbered__rail {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-top: 3px;
}

.numbered__index {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: var(--vp-c-text-3);
}

.numbered__label {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--vp-c-brand-1);
}

.numbered__claim {
  max-width: 36ch;
  margin: 0 0 10px;
  font-size: 19px;
  font-weight: 600;
  line-height: 1.42;
  letter-spacing: -0.014em;
  color: var(--vp-c-text-1);
}

.numbered__text :deep(code) {
  padding: 2px 5px;
  border-radius: 4px;
  background: var(--vp-c-default-soft);
  font-family: var(--vp-font-family-mono);
  font-size: 0.88em;
}

.numbered__text {
  max-width: 68ch;
  margin: 0;
  font-size: 15px;
  line-height: 1.7;
  color: var(--vp-c-text-2);
}

.numbered__aside {
  max-width: 66ch;
  margin: 14px 0 0;
  padding-left: 13px;
  border-left: 2px solid var(--vp-c-border);
  font-size: 13.5px;
  line-height: 1.62;
  color: var(--vp-c-text-3);
}

@media (max-width: 600px) {
  .numbered {
    grid-template-columns: 1fr;
    gap: 10px;
  }
  .numbered__rail {
    flex-direction: row;
    align-items: baseline;
    gap: 10px;
    padding-top: 0;
  }
  .numbered__claim {
    font-size: 17px;
  }
}
</style>
