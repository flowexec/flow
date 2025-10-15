---
title: MIGRATION NOTES
---

Flow Docs: Docsify → Starlight Migration

How to run locally
- cd docs
- npm install
- npm run migrate   # copies markdown to src/content/docs and assets to public
- npm run dev       # open http://localhost:4321

How to build for production
- npm run build
- Output will be in docs/dist

What was set up
- Astro + Starlight latest stable
- Dark mode enabled (theme supports system toggle)
- Pagefind search enabled (built-in Starlight)
- Sidebar structure recreated in starlight.config.ts
- GitHub integration: repo link and Edit on GitHub buttons
- Code block copy buttons enabled by default
- Static export ready for Cloudflare Pages (or any static host)

Content migration details
- All markdown is mirrored from docs/ to docs/src/content/docs/ preserving folders
- JSON schemas are copied to docs/public/schemas and available at /schemas/*
- _media assets are copied to docs/public/_media
- README.md is duplicated as index.md for the homepage
- Docsify tabs markers <!-- tabs:start/end --> are removed during migration (content remains visible but not tabbed)
- Docsify alerts (::: tip/note/warning) render as Starlight callouts without changes

Redirects and links
- /README and /README/ redirect to /
- Legacy Docsify hash URLs (/#/...) are client-only and cannot be fully redirected server-side; main entries are linked via sidebar

Checklist to verify
- Start dev server and visit homepage
- Navigate: Install, Quick Start, Guides, CLI Reference, Types, Contributing
- Verify search works (after initial build indexing)
- Verify code highlighting (bash, yaml, go, etc.) and copy buttons
- Open a JSON schema at /schemas/flowfile_schema.json
- Spot-check pages with former tabs; content should still be readable

Notes
- If you want true tabbed content, replace the removed tabs blocks with Starlight’s <Tabs> and <Tab> MDX components manually for the two identified files:
  - guide/executables.md
  - guide/secrets.md
- You can iterate on sidebar items in docs/starlight.config.ts
