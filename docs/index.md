---
layout: home

hero:
  name: "flow · open source"
  text: "Write your workflows down."
  tagline: "Then run them from any project on your machine — with the right secrets, the right environment, and a record of what happened."
  image:
    src: /icon.png
    alt: flow
  actions:
    - theme: brand
      text: Install flow
      link: /installation
    - theme: alt
      text: Quickstart
      link: /quickstart
    - theme: alt
      text: View on GitHub
      link: https://github.com/flowexec/flow
description: "Write your workflows down as flow files, then run them from any project on your machine — with encrypted secrets, execution history, and an MCP server for coding agents."
---

<NowStrip />

<p align="center"><video autoplay loop muted playsinline style="width: 100%; max-width: 900px; display: block; margin: 0 auto;"><source src="/demo/hero.mp4" type="video/mp4"><img src="/demo/hero.gif" style="width: 100%; max-width: 900px;"></video></p>

<NumberedSection index="01" label="Write" claim="Workflows are YAML you keep next to the code they run." aside="Not another YAML dialect — every flow file validates against a published JSON schema, so your editor autocompletes it before flow ever sees it.">
A flow file declares what a task does, what it needs, and how it runs: in order, in parallel, behind a condition, as an HTTP request, or as a background job. Check it in and the automation travels with the repo.
</NumberedSection>

<NumberedSection index="02" label="Run" claim="One binary, every project on your machine." aside="Not a task runner you have to cd into — flow knows where every workspace lives, so the command reads the same from any directory.">
Register a repo as a workspace and its workflows are reachable from anywhere — by name from your shell, or by browsing them in the terminal UI. Secrets come out of an encrypted local vault at the moment they are needed, never out of a file you committed.
</NumberedSection>

<NumberedSection index="03" label="Connect" claim="The same workflows, from your terminal or your coding agent." aside="Not an AI product — flow ships the tools and stops there. No vendor keys and no model calls in the critical path of a task runner.">
<code>flow mcp</code> exposes every executable over the Model Context Protocol, so an assistant can discover what a project is able to do, run it, and read the logs back — through the same engine you use by hand.
</NumberedSection>

<SectionHead icon="box" label="What's inside" moreText="All guides" moreLink="/guides/" />

<CardGrid>
  <Card icon="folder" title="Workspaces" :stack="['registry', 'git', 'discovery']" href="/guides/workspaces" alt-href="/types/workspace" alt-href-text="Schema">
    Register any repo once. Its executables are then addressable from anywhere, with that project's environment applied.
  </Card>
  <Card icon="zap" title="Executables" :stack="['serial', 'parallel', 'request', 'render', 'launch']" href="/guides/executables" alt-href="/types/flowfile" alt-href-text="Schema">
    Six execution types, conditionals, retries, arguments, and prompts — composed from other executables by reference.
  </Card>
  <Card icon="lock" title="Vaults" :stack="['AES-256', 'age', 'keyring', 'env']" href="/guides/secrets" alt-href="/cli/flow_vault" alt-href-text="CLI">
    Encrypted local secret storage with pluggable backends. Values are injected into the process at run time and never written to disk.
  </Card>
  <Card icon="file" title="Templates" :stack="['scaffolding', 'generation']" href="/guides/templating" alt-href="/types/template" alt-href-text="Schema">
    Bootstrap new projects and new workflows from templates you author, with prompts and generated files.
  </Card>
  <Card icon="monitor" title="Interactive UI" :stack="['browse', 'search', 'logs']" href="/guides/interactive" alt-href="/cli/flow_browse" alt-href-text="CLI">
    Browse, filter, and run every workflow across every workspace from one terminal interface — then watch the logs stream.
  </Card>
  <Card icon="plug" title="AI tools" badge="MCP" :stack="['Claude Code', 'Cursor', 'any MCP client']" href="/guides/ai-tools" alt-href="/cli/flow_mcp" alt-href-text="CLI">
    A deterministic tool surface for assistants: discover, run, and author workflows, with full provenance on every run.
  </Card>
</CardGrid>

<SectionHead icon="book" label="Start here" moreText="Examples gallery" moreLink="/examples" />

- **[Install flow](/installation)** — one line on macOS, Linux, or Windows
- **[Quickstart](/quickstart)** — a working workflow in about five minutes
- **[Core concepts](/guides/concepts)** — workspaces, executables, and vaults, and how they fit together
- **[Examples](/examples)** — real flow files for Go, Docker, Kubernetes, Git, and HTTP APIs
- **[flow's own automation](https://github.com/flowexec/flow/tree/main/.execs)** — this project builds, tests, and releases itself with flow
- **[How flow is built](https://jahvon.dev/architecture/flow/)** — an architecture deep dive from Dockery Labs, the workshop behind flow

::: tip Mochi: your workflows, neatly wrapped
Want a desktop dashboard, workflow auto-discovery, and AI enrichment on top of flow?
[Sign up for early access →](https://mochiexec.io)
:::

<p class="home-closer">A short path to everything you run.</p>
