import { defineConfig } from 'vitepress'
import { tabsMarkdownPlugin } from 'vitepress-plugin-tabs'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "flow",
  description: "Local developer automation platform that flows with you.",
  base: '/',
  outDir: './dist',

  head: [
    ['link', { rel: 'icon', href: '/favicon.ico' }]
  ],

  markdown: {
    config(md) {
      md.use(tabsMarkdownPlugin)
    },
  },

  themeConfig: {
    logo: {
      light: '/logo-light.png',
      dark: '/logo-dark.png'
    },

    siteTitle: false,

    nav: [
      { text: 'Home', link: '/' },
      { text: 'Guides', link: '/guides/', activeMatch: '/guides/'},
      { text: 'CLI Reference', link: '/cli/', activeMatch: '/cli/' },
      { text: 'Config Reference', link: '/types/', activeMatch: '/types/' }
    ],

    sidebar: {
      '/guides/': [
        {
          text: 'User Guides',
          items: [
            { text: 'Overview', link: '/guides/' },
            { text: 'Getting Started',
              items: [
                { text: 'Concepts', link: '/guides/concepts' },
                { text: 'Your First Workflow', link: '/guides/first-workflow' },
              ]
            },
            { text: 'Essentials',
              items: [
                { text: 'Executables', link: '/guides/executables' },
                { text: 'Workspaces', link: '/guides/workspaces' },
                { text: 'Secrets', link: '/guides/secrets' },
            ]},
            { text: 'Advanced',
            items: [
              { text: 'Imported Executables', link: '/guides/generated-config' },
              { text: 'Templates & Workflow Generation', link: '/guides/templating' },
              { text: 'Advanced Workflows', link: '/guides/advanced' },
              { text: 'Interactive UI', link: '/guides/interactive' },
              { text: 'Integrations', link: '/guides/integrations' },
            ]},
          ]
        }
      ],
      '/cli/': [
        {
          text: 'CLI Reference',
          items: [
            { text: 'Overview', link: '/cli/flow' },
            { text: 'flow browse', link: '/cli/flow_browse' },
            { text: 'flow exec', link: '/cli/flow_exec' },
            { text: 'flow logs', link: '/cli/flow_logs' },
            { text: 'flow mcp', link: '/cli/flow_mcp' },
            { text: 'flow sync', link: '/cli/flow_sync' },
            {
              text: 'Cache',
              collapsed: true,
              items: [
                { text: 'flow cache', link: '/cli/flow_cache' },
                { text: 'flow cache clear', link: '/cli/flow_cache_clear' },
                { text: 'flow cache get', link: '/cli/flow_cache_get' },
                { text: 'flow cache list', link: '/cli/flow_cache_list' },
                { text: 'flow cache remove', link: '/cli/flow_cache_remove' },
                { text: 'flow cache set', link: '/cli/flow_cache_set' }
              ]
            },
            {
              text: 'Config',
              collapsed: true,
              items: [
                { text: 'flow config', link: '/cli/flow_config' },
                { text: 'flow config get', link: '/cli/flow_config_get' },
                { text: 'flow config reset', link: '/cli/flow_config_reset' },
                { text: 'flow config set', link: '/cli/flow_config_set' },
                { text: 'flow config set log-mode', link: '/cli/flow_config_set_log-mode' },
                { text: 'flow config set namespace', link: '/cli/flow_config_set_namespace' },
                { text: 'flow config set notifications', link: '/cli/flow_config_set_notifications' },
                { text: 'flow config set theme', link: '/cli/flow_config_set_theme' },
                { text: 'flow config set timeout', link: '/cli/flow_config_set_timeout' },
                { text: 'flow config set tui', link: '/cli/flow_config_set_tui' },
                { text: 'flow config set workspace', link: '/cli/flow_config_set_workspace' },
                { text: 'flow config set workspace-mode', link: '/cli/flow_config_set_workspace-mode' }
              ]
            },
            {
              text: 'Secret',
              collapsed: true,
              items: [
                { text: 'flow secret', link: '/cli/flow_secret' },
                { text: 'flow secret get', link: '/cli/flow_secret_get' },
                { text: 'flow secret list', link: '/cli/flow_secret_list' },
                { text: 'flow secret remove', link: '/cli/flow_secret_remove' },
                { text: 'flow secret set', link: '/cli/flow_secret_set' }
              ]
            },
            {
              text: 'Template',
              collapsed: true,
              items: [
                { text: 'flow template', link: '/cli/flow_template' },
                { text: 'flow template add', link: '/cli/flow_template_add' },
                { text: 'flow template generate', link: '/cli/flow_template_generate' },
                { text: 'flow template get', link: '/cli/flow_template_get' },
                { text: 'flow template list', link: '/cli/flow_template_list' }
              ]
            },
            {
              text: 'Vault',
              collapsed: true,
              items: [
                { text: 'flow vault', link: '/cli/flow_vault' },
                { text: 'flow vault create', link: '/cli/flow_vault_create' },
                { text: 'flow vault edit', link: '/cli/flow_vault_edit' },
                { text: 'flow vault get', link: '/cli/flow_vault_get' },
                { text: 'flow vault list', link: '/cli/flow_vault_list' },
                { text: 'flow vault remove', link: '/cli/flow_vault_remove' },
                { text: 'flow vault switch', link: '/cli/flow_vault_switch' }
              ]
            },
            {
              text: 'Workspace',
              collapsed: true,
              items: [
                { text: 'flow workspace', link: '/cli/flow_workspace' },
                { text: 'flow workspace add', link: '/cli/flow_workspace_add' },
                { text: 'flow workspace get', link: '/cli/flow_workspace_get' },
                { text: 'flow workspace list', link: '/cli/flow_workspace_list' },
                { text: 'flow workspace remove', link: '/cli/flow_workspace_remove' },
                { text: 'flow workspace switch', link: '/cli/flow_workspace_switch' },
                { text: 'flow workspace view', link: '/cli/flow_workspace_view' }
              ]
            }
          ]
        }
      ],
      '/types/': [
        {
          text: 'Configuration Reference',
          items: [
            { text: 'Overview', link: '/types/' },
            { text: 'Config', link: '/types/config' },
            { text: 'Flow File', link: '/types/flowfile' },
            { text: 'Template', link: '/types/template' },
            { text: 'Workspace', link: '/types/workspace' }
          ]
        }
      ],
      '/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Installation', link: '/installation' },
            { text: 'Quick Start', link: '/quickstart' }
          ]
        },
        {
          text: 'More',
          items: [
            { text: 'User Guides', link: '/guides/' },
            { text: 'CLI Reference', link: '/cli/' },
            { text: 'Configuration Reference', link: '/types/' },
            { text: 'Contributing', link: '/development' },
            { text: 'TUI Kit', link: '/tuikit' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/flowexec/flow' },
      { icon: 'discord', link: 'https://discord.gg/CtByNKNMxM' }
    ],

    search: {
      provider: 'local'
    },

    outline: {
      level: [2, 3]
    }
  }
})
