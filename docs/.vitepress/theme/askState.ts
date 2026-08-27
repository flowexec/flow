import { ref } from 'vue'

/**
 * Shared state for the Ask panel. The trigger renders in two places (desktop
 * nav and mobile nav screen) but the modal is mounted once by Layout.vue, so
 * the open flag has to live outside both.
 */
export const askOpen = ref(false)
export const askQuery = ref('')

export function openAsk(query = '') {
  if (query) askQuery.value = query
  askOpen.value = true
}
