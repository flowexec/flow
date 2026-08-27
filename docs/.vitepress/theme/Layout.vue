<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, onUnmounted } from 'vue'
import DefaultTheme from 'vitepress/theme'
import { useData } from 'vitepress'
import AskTrigger from './AskTrigger.vue'
import AskBridge from './AskBridge.vue'
import { askOpen, openAsk } from './askState'

const { Layout } = DefaultTheme
const { frontmatter } = useData()

// Both of these pull in weight the rest of the site never needs — markdown-it
// for the answer, a canvas loop for the backdrop — so neither is in the main
// chunk.
const AskDeepWiki = defineAsyncComponent(() => import('./AskDeepWiki.vue'))
const HomeBackdrop = defineAsyncComponent(() => import('./HomeBackdrop.vue'))

const isHome = computed(() => frontmatter.value.layout === 'home')

function isEditingContent(event: KeyboardEvent) {
  const el = event.target as HTMLElement | null
  return (
    !!el &&
    (el.isContentEditable ||
      el.tagName === 'INPUT' ||
      el.tagName === 'SELECT' ||
      el.tagName === 'TEXTAREA')
  )
}

// VitePress already owns ⌘K and `/` for local search.
function onKeydown(event: KeyboardEvent) {
  if (event.key.toLowerCase() !== 'i' || !(event.metaKey || event.ctrlKey)) return
  if (isEditingContent(event)) return
  event.preventDefault()
  openAsk()
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Layout>
    <!-- Before the menu, so the two ways to look something up sit together
         instead of at opposite ends of the bar. -->
    <template #nav-bar-content-before>
      <AskTrigger />
    </template>
    <template #nav-screen-content-after>
      <AskTrigger />
    </template>
  </Layout>

  <HomeBackdrop v-if="isHome" />
  <AskBridge />
  <AskDeepWiki v-if="askOpen" />
</template>
