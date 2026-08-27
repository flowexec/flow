<script setup lang="ts">
import { computed } from 'vue'
import { ICONS, FILLED } from './icons'

const props = defineProps<{ name: string }>()

const markup = computed(() => ICONS[props.name] ?? '')
const filled = computed(() => FILLED.has(props.name))
</script>

<template>
  <svg
    class="ui-icon"
    viewBox="0 0 24 24"
    :fill="filled ? 'currentColor' : 'none'"
    :stroke="filled ? 'none' : 'currentColor'"
    stroke-width="1.6"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    focusable="false"
    v-html="markup"
  />
</template>

<style scoped>
.ui-icon {
  width: 1em;
  height: 1em;
  flex: 0 0 auto;
  opacity: 0.75;
}
</style>
