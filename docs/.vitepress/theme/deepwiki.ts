/**
 * DeepWiki client.
 *
 * DeepWiki indexes flowexec/flow and exposes an `ask_question` tool over a
 * public, unauthenticated MCP endpoint. The endpoint is stateless — no
 * `initialize` handshake and no session id — so a single POST is the whole
 * protocol, and it answers `Access-Control-Allow-Origin: *`, which is what
 * makes calling it straight from the page possible.
 *
 * It replies with an SSE stream: progress notifications while the model works
 * (~15-20s is normal), then one frame carrying the answer.
 *
 * None of this is versioned or documented by DeepWiki. Callers must handle
 * failure by pointing people at deepwiki.com rather than showing a dead UI.
 */

const ENDPOINT = 'https://mcp.deepwiki.com/mcp'
export const REPO = 'flowexec/flow'
export const DEEPWIKI_URL = `https://deepwiki.com/${REPO}`

/** Generous ceiling: observed round trips are ~16s. */
const TIMEOUT_MS = 90_000

export interface AskOptions {
  /** Seconds elapsed, as reported by DeepWiki's own progress notifications. */
  onProgress?: (elapsedSeconds: number) => void
  signal?: AbortSignal
}

export class DeepWikiError extends Error {}

export async function askDeepWiki(
  question: string,
  { onProgress, signal }: AskOptions = {}
): Promise<string> {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), TIMEOUT_MS)
  const relay = () => controller.abort()
  signal?.addEventListener('abort', relay)

  try {
    const res = await fetch(ENDPOINT, {
      method: 'POST',
      // Only `content-type` is in the endpoint's Access-Control-Allow-Headers;
      // `accept` rides along because it is a CORS-safelisted request header.
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json, text/event-stream',
      },
      body: JSON.stringify({
        jsonrpc: '2.0',
        id: 1,
        method: 'tools/call',
        params: {
          name: 'ask_question',
          arguments: { repoName: REPO, question },
        },
      }),
      signal: controller.signal,
    })

    if (!res.ok || !res.body) {
      throw new DeepWikiError(`DeepWiki returned ${res.status}`)
    }

    const answer = await readStream(res.body, onProgress)
    if (!answer) {
      throw new DeepWikiError('DeepWiki closed the stream without an answer')
    }
    return tidy(answer)
  } finally {
    clearTimeout(timeout)
    signal?.removeEventListener('abort', relay)
  }
}

async function readStream(
  body: ReadableStream<Uint8Array>,
  onProgress?: (seconds: number) => void
): Promise<string | null> {
  const reader = body.pipeThrough(new TextDecoderStream()).getReader()
  let buffer = ''
  let answer: string | null = null

  const handle = (frame: string) => {
    const payload = parseFrame(frame)
    if (!payload) return

    if (payload.error) {
      throw new DeepWikiError(payload.error.message ?? 'DeepWiki error')
    }

    if (payload.method === 'notifications/message') {
      const msg = payload.params?.data?.msg
      const seconds = typeof msg === 'string' ? elapsedFrom(msg) : null
      if (seconds !== null) onProgress?.(seconds)
      return
    }

    const text = payload.result?.content?.[0]?.text
    if (typeof text === 'string') answer = text
  }

  try {
    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      buffer += value

      // DeepWiki terminates frames with CRLFCRLF, not the bare LFLF that most
      // SSE examples show. Hold the trailing partial back for the next chunk.
      const frames = buffer.split(/\r?\n\r?\n/)
      buffer = frames.pop() ?? ''
      frames.forEach(handle)
    }
    // The answer is the last thing sent and may arrive without a closing blank
    // line, so whatever is left when the stream ends still has to be read.
    if (buffer.trim()) handle(buffer)
  } finally {
    reader.cancel().catch(() => {})
  }

  return answer
}

interface Frame {
  method?: string
  params?: { data?: { msg?: string } }
  result?: { content?: Array<{ type?: string; text?: string }> }
  error?: { message?: string }
}

function parseFrame(frame: string): Frame | null {
  for (const line of frame.split('\n')) {
    // `:` lines are keep-alive comments.
    if (!line.startsWith('data:')) continue
    try {
      return JSON.parse(line.slice(5).trim()) as Frame
    } catch {
      return null
    }
  }
  return null
}

function elapsedFrom(msg: string): number | null {
  const match = /\((\d+)s elapsed\)/.exec(msg)
  return match ? Number(match[1]) : null
}

/**
 * DeepWiki strips its inline citation markers on the way out and leaves the
 * whitespace behind, so answers arrive with runs of spaces and orphaned gaps
 * before punctuation. Fenced code is left exactly as sent.
 */
function tidy(markdown: string): string {
  return markdown
    .split(/(```[\s\S]*?```)/g)
    .map((segment, i) =>
      i % 2 === 1
        ? segment
        : segment
            .replace(/[ \t]+([.,;:!?)])/g, '$1')
            .replace(/[ \t]{2,}/g, ' ')
            .replace(/[ \t]+$/gm, '')
            .replace(/\]\(\/(wiki|search)\//g, '](https://deepwiki.com/$1/')
    )
    .join('')
    .trim()
}
