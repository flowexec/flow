---
layout: home

hero:
  name: "flow · open source"
  text: "Write your workflows down."
  tagline: "Then run them from any project on your machine, with the right secrets and a record of every run."
  image:
    src: /icon.png
    alt: flow
  actions:
    - theme: brand
      text: What is flow?
      link: /introduction/what-is-flow
    - theme: alt
      text: Quickstart
      link: /quickstart
    - theme: alt
      text: View on GitHub
      link: https://github.com/flowexec/flow
description: "flow is an open-source workflow engine. Write tasks down as YAML flow files, run them from any project with secrets from a local vault, and keep a record of every run, including your coding agents'."
---

<NowStrip />

<p align="center"><video autoplay loop muted playsinline style="width: 100%; max-width: 900px; display: block; margin: 0 auto;"><source src="/demo/hero.mp4" type="video/mp4"><img src="/demo/hero.gif" alt="flow browse listing executables across workspaces and running one" style="width: 100%; max-width: 900px;"></video></p>

<SectionHead icon="box" label="Guides" moreText="All guides" moreLink="/guides/" />

<CardGrid>
  <Card icon="folder" title="Workspaces" :stack="['local folders', 'Git URLs']" href="/guides/workspaces" alt-href="/types/workspace" alt-href-text="Schema">
    Register a repo once and run its tasks by name from any directory.
  </Card>
  <Card icon="zap" title="Executables" :stack="['exec', 'serial', 'parallel', 'request', 'render', 'launch']" href="/guides/executables" alt-href="/types/flowfile" alt-href-text="Schema">
    A task declares what it runs and what it needs. Build bigger ones from smaller ones by reference.
  </Card>
  <Card icon="lock" title="Vaults" :stack="['AES-256', 'age', 'keyring', 'env']" href="/guides/secrets" alt-href="/cli/flow_vault" alt-href-text="CLI">
    Secrets stay encrypted on your machine and reach a task only while it runs.
  </Card>
  <Card icon="file" title="Templates" :stack="['prompts', 'generated files']" href="/guides/templating" alt-href="/types/template" alt-href-text="Schema">
    Start a new project or workflow from a template you wrote.
  </Card>
  <Card icon="monitor" title="Terminal UI" :stack="['flow browse', 'flow logs']" href="/guides/interactive" alt-href="/cli/flow_browse" alt-href-text="CLI">
    Find and run any task across every workspace, then follow its output.
  </Card>
  <Card icon="plug" title="Coding agents" badge="MCP" :stack="['Claude Code', 'Cursor', 'any MCP client']" href="/guides/ai-tools" alt-href="/cli/flow_mcp" alt-href-text="CLI">
    Agents run the same tasks you do, and each run records which agent started it.
  </Card>
</CardGrid>

<SectionHead icon="book" label="Start here" />

- **[Coming from Make, npm, just or Task](/guides/coming-from)**: what changes and what doesn't
- **[Examples](/examples)**: flow files for Go, Docker, Kubernetes, Git and HTTP APIs
- **[CLI reference](/cli/)**: every `flow` command and flag
- **[flow's own automation](https://github.com/flowexec/flow/tree/main/.execs)**: this project builds, tests and releases itself with flow
- **[How flow is built](https://jahvon.dev/architecture/flow/)**: an architecture deep dive

<MochiRef placement="home" eyebrow="New · Mochi" title="Mochi is out: the desktop app for flow">

It opens your flow workspaces as they are, finds the Makefiles and npm scripts in your other repos, and adds run search and controls over what agents may run. Free for 7 days, no card.

</MochiRef>

<p class="home-closer">A short path <span class="muted">to</span> everything you run.</p>
