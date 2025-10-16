import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import remarkFixMdLinks from './remark/remark-fix-md-links.mjs';
import starlightThemeFlexoki from 'starlight-theme-flexoki'

// https://docs.astro.build/en/reference/configuration-reference/
export default defineConfig({
  site: 'https://flowexec.io',
  base: '/',
  output: 'static',
  markdown: {
    remarkPlugins: [remarkFixMdLinks],
  },
  integrations: [
    // Use Starlight with configuration from starlight.config.ts
    starlight({
        title: 'flow',
        lastUpdated: false,
        social: [
            {icon: 'github', label: 'GitHub', href: 'https://github.com/flowexec/flow'},
            {icon: 'discord', label: 'Discord', href: 'https://discord.gg/CtByNKNMxM'}
        ],
        favicon: '/assets/favicon.ico',
        logo: {
            light: './src/assets/logo-light.png',
            dark: './src/assets/logo-dark.png',
            replacesTitle: true,
        },
        plugins: [
            starlightThemeFlexoki({accentColor: "magenta"}),
        ],
        components: { },
        tableOfContents: { minHeadingLevel: 2, maxHeadingLevel: 3 },
        sidebar: [
            {label: "Install", link: "installation"},
            {label: "Quick Start", link: "quickstart"},
            {label: "User Guides", collapsed: true, autogenerate: {directory: "guides"}},
            {label: "CLI Reference", collapsed: true, autogenerate: {directory: "cli"}},
            {label: "Configuration Reference", collapsed: true, autogenerate: {directory: "types"}},
            {label: "Contributing", link: "development"},
        ]
    }),
  ],
});
