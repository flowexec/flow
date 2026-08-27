import { defineConfig } from 'vitepress'
import { tabsMarkdownPlugin } from 'vitepress-plugin-tabs'
import flowfileGrammar from './flowfile.tmLanguage.json'

// https://vitepress.dev/reference/site-config
const SITE = 'https://flowexec.io'
const SITE_NAME = 'flow'
const SITE_DESCRIPTION =
  'Write your workflows down, then run them from any project on your machine — with the right secrets, the right environment, and a record of what happened.'
const OG_IMAGE = `${SITE}/og-default.png`

export default defineConfig({
  title: SITE_NAME,
  description: SITE_DESCRIPTION,
  base: '/',
  outDir: './dist',
  lang: 'en-US',

  // Cloudflare Pages redirects /foo.html to /foo, so without this the sitemap
  // advertises URLs that immediately redirect and the site links to a different
  // form than it declares canonical.
  cleanUrls: true,

  sitemap: {
    hostname: SITE
  },

  head: [
    ['link', { rel: 'icon', href: '/favicon.ico', sizes: '48x48' }],
    ['link', { rel: 'icon', type: 'image/png', href: '/icon.png' }],
    ['link', { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' }],
    ['meta', { name: 'theme-color', content: '#2D353B' }],
    ['meta', { name: 'author', content: 'Dockery Labs' }],
    ['link', { rel: 'alternate', type: 'text/plain', title: 'llms.txt', href: `${SITE}/llms.txt` }],

    // Per-page og:title / og:description / og:url and the canonical link are
    // added in transformPageData below; these are the values that never vary.
    ['meta', { property: 'og:site_name', content: SITE_NAME }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:locale', content: 'en_US' }],
    ['meta', { property: 'og:image', content: OG_IMAGE }],
    ['meta', { property: 'og:image:width', content: '1200' }],
    ['meta', { property: 'og:image:height', content: '630' }],
    ['meta', { property: 'og:image:alt', content: 'flow' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:image', content: OG_IMAGE }],

    ['script', { type: 'application/ld+json' }, JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      name: SITE_NAME,
      description: SITE_DESCRIPTION,
      url: SITE,
      applicationCategory: 'DeveloperApplication',
      operatingSystem: 'macOS, Linux, Windows',
      license: 'https://github.com/flowexec/flow/blob/main/LICENSE',
      offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD' },
      author: { '@type': 'Organization', name: 'Dockery Labs', url: 'https://jahvon.dev' },
      sameAs: ['https://github.com/flowexec/flow', 'https://discord.gg/CtByNKNMxM']
    })]
  ],

  transformPageData(pageData) {
    const path = pageData.relativePath
      .replace(/(^|\/)index\.md$/, '$1')
      .replace(/\.md$/, '')
    const canonical = `${SITE}/${path}`
    const title = pageData.frontmatter.title
      ? `${pageData.frontmatter.title} | ${SITE_NAME}`
      : SITE_NAME
    const description = pageData.frontmatter.description || SITE_DESCRIPTION

    pageData.frontmatter.head ??= []
    pageData.frontmatter.head.push(
      ['link', { rel: 'canonical', href: canonical }],
      ['meta', { property: 'og:url', content: canonical }],
      ['meta', { property: 'og:title', content: title }],
      ['meta', { property: 'og:description', content: description }],
      ['meta', { name: 'twitter:title', content: title }],
      ['meta', { name: 'twitter:description', content: description }]
    )
  },

  markdown: {
    config(md) {
      md.use(tabsMarkdownPlugin)
    },
    languages: [flowfileGrammar as never],
    // Shiki bundles both, so code blocks land on the same palette as the rest
    // of the site instead of GitHub's default blues.
    theme: {
      light: 'everforest-light',
      dark: 'everforest-dark',
    },
  },

  themeConfig: {
    logo: {
      light: '/logo-light.png',
      dark: '/logo-dark.png'
    },

    siteTitle: false,

    // Three items, not five. The logo already goes home, and the two reference
    // sections are destinations you arrive at from a guide rather than things
    // you browse cold — so they collapse into one menu and give the search and
    // Ask controls room to breathe.
    nav: [
      { text: 'Guides', link: '/guides/', activeMatch: '/guides/' },
      { text: 'Examples', link: '/examples', activeMatch: '/examples' },
      {
        text: 'Reference',
        activeMatch: '/(cli|types)/',
        items: [
          { text: 'CLI Reference', link: '/cli/' },
          { text: 'Configuration Reference', link: '/types/' },
          { text: 'Contributing', link: '/development' }
        ]
      }
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
                { text: 'History & Logs', link: '/guides/execution-history' },
              ]
            },
            { text: 'Advanced',
              items: [
                { text: 'Expression Language', link: '/guides/expressions' },
                { text: 'Advanced Workflows', link: '/guides/advanced' },
                { text: 'Imported Executables', link: '/guides/generated-config' },
                { text: 'Templates & Workflow Generation', link: '/guides/templating' },
              ]
            },
            { text: 'Interfaces',
              items: [
                { text: 'Interactive UI', link: '/guides/interactive' },
                { text: 'Run Provenance', link: '/guides/run-provenance' },
              ]
            },
            { text: 'Integrations',
              items: [
                { text: 'AI Tools & MCP', link: '/guides/ai-tools' },
                { text: 'Containers', link: '/guides/containers' },
                { text: 'GitHub Actions', link: '/guides/github-actions' },
              ]
            },
          ]
        }
      ],
      '/cli/': [
        {
          text: 'CLI Reference',
          items: [
            { text: 'Overview', link: '/cli/flow' },
            {
              text: 'Core',
              items: [
                { text: 'flow browse', link: '/cli/flow_browse' },
                { text: 'flow exec', link: '/cli/flow_exec' },
                { text: 'flow mcp', link: '/cli/flow_mcp' },
                { text: 'flow sync', link: '/cli/flow_sync' },
              ]
            },
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
                { text: 'flow config set update-check', link: '/cli/flow_config_set_update-check' },
                { text: 'flow config set workspace', link: '/cli/flow_config_set_workspace' },
                { text: 'flow config set workspace-mode', link: '/cli/flow_config_set_workspace-mode' }
              ]
            },
            {
              text: 'Logs',
              collapsed: true,
              items: [
                { text: 'flow logs', link: '/cli/flow_logs' },
                { text: 'flow logs attach', link: '/cli/flow_logs_attach' },
                { text: 'flow logs clear', link: '/cli/flow_logs_clear' },
                { text: 'flow logs kill', link: '/cli/flow_logs_kill' }
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
                { text: 'flow template list', link: '/cli/flow_template_list' },
                { text: 'flow template remove', link: '/cli/flow_template_remove' }
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
                { text: 'flow workspace update', link: '/cli/flow_workspace_update' },
                { text: 'flow workspace view', link: '/cli/flow_workspace_view' }
              ]
            },
            {
              text: 'Tools',
              collapsed: true,
              items: [
                { text: 'flow cli', link: '/cli/flow_cli' },
                { text: 'flow cli update', link: '/cli/flow_cli_update' },
                { text: 'flow schema', link: '/cli/flow_schema' },
                { text: 'flow schema validate', link: '/cli/flow_schema_validate' }
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
            { text: 'Quick Start', link: '/quickstart' },
            { text: 'Examples', link: '/examples' }
          ]
        },
        {
          text: 'More',
          items: [
            { text: 'User Guides', link: '/guides/' },
            { text: 'CLI Reference', link: '/cli/' },
            { text: 'Configuration Reference', link: '/types/' },
            { text: 'Contributing', link: '/development' },
            { text: 'Breaking Changes', link: '/breaking-changes' },
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

    footer: {
      message: 'Released under the <a href="https://github.com/flowexec/flow/blob/main/LICENSE">Apache 2.0 License</a>.',
      copyright: [
        `&copy; ${new Date().getFullYear()} <a href="https://jahvon.dev">Dockery Labs</a>`,
        '<a href="https://jahvon.dev/architecture/flow/">Architecture</a>',
        '<a href="https://mochiexec.io">Mochi</a>',
      ].join(' &middot; ')
    },

    outline: {
      level: [2, 3]
    }
  }
})
