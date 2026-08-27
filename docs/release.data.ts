import { defineLoader } from 'vitepress'

const REPO = 'flowexec/flow'
const GITHUB_API = 'https://api.github.com'

export interface ReleaseData {
  /** Latest release tag, e.g. "v1.4.2". Empty when the lookup failed. */
  tag: string
  url: string
}

declare const data: ReleaseData
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

const RELEASES_URL = `https://github.com/${REPO}/releases`

export default defineLoader({
  async load(): Promise<ReleaseData> {
    // Unlike examples.data.ts this never throws: the version is decoration, and
    // an unauthenticated CI build that trips GitHub's rate limit should still
    // produce a site.
    try {
      const res = await fetch(`${GITHUB_API}/repos/${REPO}/releases/latest`, {
        headers: githubHeaders(),
      })
      if (!res.ok) {
        console.warn(`[release.data] GitHub API ${res.status}; omitting version`)
        return { tag: '', url: RELEASES_URL }
      }
      const release = (await res.json()) as { tag_name?: string; html_url?: string }
      return {
        tag: release.tag_name ?? '',
        url: release.html_url ?? RELEASES_URL,
      }
    } catch (err) {
      console.warn(`[release.data] lookup failed; omitting version:`, err)
      return { tag: '', url: RELEASES_URL }
    }
  },
})
