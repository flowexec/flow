import { defineConfig, type Sidebar } from '@astrojs/starlight/config';

const sidebar: Sidebar = [
  {
    label: 'Getting Started',
    items: [
      { label: 'Install', link: '/installation/' },
      { label: 'Quick Start', link: '/quickstart/' },
    ],
  },
  {
    label: 'Guides',
    items: [
      { label: 'Overview', link: '/guide/README/' },
      { label: 'Core Concepts', link: '/guide/concepts/' },
      { label: 'Your First Workflow', link: '/guide/first-workflow/' },
      { label: 'Workspaces', link: '/guide/workspaces/' },
      { label: 'Executables', link: '/guide/executables/' },
      { label: 'Working with Secrets', link: '/guide/secrets/' },
      { label: 'Templates & Workflow Generation', link: '/guide/templating/' },
      { label: 'Advanced Workflows', link: '/guide/advanced/' },
      { label: 'Interactive UI', link: '/guide/interactive/' },
      { label: 'Integrations', link: '/guide/integrations/' },
    ],
  },
  {
    label: 'CLI Reference',
    items: [
      { label: 'Overview', link: '/cli/README/' },
      { label: 'flow exec', link: '/cli/flow_exec/' },
      { label: 'flow browse', link: '/cli/flow_browse/' },
      { label: 'flow template', link: '/cli/flow_template/' },
      { label: 'flow config', link: '/cli/flow_config/' },
      { label: 'flow workspace', link: '/cli/flow_workspace/' },
      { label: 'flow secret', link: '/cli/flow_secret/' },
      { label: 'flow vault', link: '/cli/flow_vault/' },
      { label: 'flow sync', link: '/cli/flow_sync/' },
      { label: 'flow logs', link: '/cli/flow_logs/' },
    ],
  },
  {
    label: 'Configuration Reference',
    items: [
      { label: 'Overview', link: '/types/README/' },
      { label: 'flow file (executables)', link: '/types/flowfile/' },
      { label: 'template file', link: '/types/template/' },
      { label: 'workspace file', link: '/types/workspace/' },
      { label: 'user configuration file', link: '/types/config/' },
    ],
  },
  { label: 'Contributing', link: '/development/' },
];

export default defineConfig({
  title: 'flow',
  lastUpdated: true,
  social: { github: 'https://github.com/flowexec/flow' },
  favicon: '/_media/favicon.ico',
  components: { },
  sidebar,
  tableOfContents: { minHeadingLevel: 2, maxHeadingLevel: 4 },
  shikiConfig: { themes: { light: 'github-light', dark: 'github-dark' } },
  editLink: {
    baseUrl: 'https://github.com/flowexec/flow/edit/main/docs/src/content/docs/'
  },
  site: 'https://flowexec.github.io/flow',
});
