import { defineLoader, createMarkdownRenderer } from 'vitepress'
import { load as yamlLoad } from 'js-yaml'
import { fileURLToPath } from 'url'
import path from 'path'
import flowfileGrammar from './.vitepress/flowfile.tmLanguage.json'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const REPO = 'flowexec/examples'
const BRANCH = 'main'
const GITHUB_API = 'https://api.github.com'
const GITHUB_RAW = 'https://raw.githubusercontent.com'

export interface FlowExecutable {
  verb: string
  name?: string
  description?: string
}

export interface FlowFile {
  path: string
  category: string
  filename: string
  namespace: string
  description: string
  tags: string[]
  executables: FlowExecutable[]
  rawContent: string
  highlightedContent: string
  sourceUrl: string
}

export interface ExamplesData {
  categories: string[]
  files: FlowFile[]
}

declare const data: ExamplesData
export { data }

function githubHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    Accept: 'application/vnd.github+json',
    'X-GitHub-Api-Version': '2022-11-28',
  }
  if (process.env.GITHUB_TOKEN) {
    headers['Authorization'] = `Bearer ${process.env.GITHUB_TOKEN}`
  }
  return headers
}

export default defineLoader({
  async load(): Promise<ExamplesData> {
    const md = await createMarkdownRenderer(__dirname, {
      languages: [flowfileGrammar as never],
    })

    const treeRes = await fetch(
      `${GITHUB_API}/repos/${REPO}/git/trees/${BRANCH}?recursive=1`,
      { headers: githubHeaders() }
    )
    if (!treeRes.ok) {
      throw new Error(`GitHub API error ${treeRes.status}: ${await treeRes.text()}`)
    }
    const tree = (await treeRes.json()) as {
      tree: Array<{ path: string; type: string }>
    }

    const flowPaths = tree.tree
      .filter(
        (f) =>
          f.type === 'blob' &&
          (f.path.endsWith('.flow') || f.path.endsWith('.flow.yaml')) &&
          !f.path.endsWith('.flow.tmpl') &&
          f.path !== 'validate.flow' &&
          f.path !== 'validate.flow.yaml'
      )
      .map((f) => f.path)

    const files = await Promise.all(
      flowPaths.map(async (filePath) => {
        const raw = await fetch(
          `${GITHUB_RAW}/${REPO}/${BRANCH}/${filePath}`,
          { headers: githubHeaders() }
        )
        if (!raw.ok) {
          throw new Error(`Failed to fetch ${filePath}: ${raw.status}`)
        }
        const content = await raw.text()

        let parsed: Record<string, unknown> = {}
        try {
          parsed = (yamlLoad(content) as Record<string, unknown>) ?? {}
        } catch {
          // leave parsed empty on invalid YAML
        }

        const parts = filePath.split('/')
        const filename = parts[parts.length - 1]
        const baseName = filename.replace(/\.flow(\.yaml)?$/, '')

        // Derive category: subdirectory name if nested, else infer from tags
        let category: string
        if (parts.length > 1) {
          category = parts[0]
        } else {
          const KNOWN = ['basics', 'go-project', 'docker', 'setup', 'git', 'api', 'kubernetes']
          const execTags = ((parsed.executables as Array<Record<string, unknown>>) ?? [])
            .flatMap((e) => (e.tags as string[]) ?? [])
          category = KNOWN.find((c) => execTags.includes(c)) ?? 'general'
        }

        // Merge file-level and executable-level tags for older flat structure
        const fileTags = (parsed.tags as string[]) ?? []
        const execTags = ((parsed.executables as Array<Record<string, unknown>>) ?? [])
          .flatMap((e) => (e.tags as string[]) ?? [])
        const tags = [...new Set([...fileTags, ...execTags])].sort()

        const highlighted = md.render('```flowfile\n' + content + '\n```')

        const executables = (
          (parsed.executables as Array<Record<string, unknown>>) ?? []
        ).map((e) => ({
          verb: e.verb as string,
          name: e.name as string | undefined,
          description: ((e.description as string) ?? '').trim(),
        }))

        return {
          path: filePath,
          category,
          filename,
          namespace: (parsed.namespace as string) ?? baseName,
          description: ((parsed.description as string) ?? '').trim(),
          tags,
          executables,
          rawContent: content,
          highlightedContent: highlighted,
          sourceUrl: `https://github.com/${REPO}/blob/${BRANCH}/${filePath}`,
        }
      })
    )

    files.sort((a, b) => {
      if (a.category !== b.category) return a.category.localeCompare(b.category)
      return a.filename.localeCompare(b.filename)
    })

    const categories = [...new Set(files.map((f) => f.category))].sort()
    return { categories, files }
  },
})
