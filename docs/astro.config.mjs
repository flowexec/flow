import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import remarkFixMdLinks from './remark/remark-fix-md-links.mjs';

// https://docs.astro.build/en/reference/configuration-reference/
export default defineConfig({
  site: 'https://flowexec.io',
  base: '/',
  output: 'static',
  markdown: {
    remarkPlugins: [remarkFixMdLinks],
  },
  redirects: {
    '/README': '/',
    '/README/': '/',
  },
  integrations: [
    // Use Starlight with configuration from starlight.config.ts
    starlight({
        title: 'flow',
        lastUpdated: true,
        social: {
            discord: 'https://discord.gg/CtByNKNMxM',
            github: 'https://github.com/flowexec/flow',
        },
        favicon: '/assets/favicon.ico',
        logo: {
            light: './src/assets/logo-light.png',
            dark: './src/assets/logo-dark.png',
            replacesTitle: true,
        },
        components: { },
        tableOfContents: { minHeadingLevel: 2, maxHeadingLevel: 4 },
        // shikiConfig: { themes: { light: 'github-light', dark: 'github-dark' } },
        editLink: {
            baseUrl: 'https://github.com/flowexec/flow/edit/main/docs/src/content/docs/'
        },
    }),
  ],
});
