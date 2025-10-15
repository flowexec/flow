#!/usr/bin/env node
import { promises as fs } from 'fs';
import path from 'path';

const repoRoot = path.resolve(process.cwd());
const docsRoot = repoRoot; // script is intended to run from docs/ directory
const srcDir = path.join(docsRoot, 'src', 'content', 'docs');
const publicDir = path.join(docsRoot, 'public');

const IGNORE_DIRS = new Set([
  'node_modules', 'src', 'dist', '.astro', '.git', '.next', 'public'
]);

async function ensureDir(dir) {
  await fs.mkdir(dir, { recursive: true });
}

async function copyFilePreserveDirs(from, toRoot) {
  const rel = path.relative(docsRoot, from);
  const to = path.join(toRoot, rel);
  await ensureDir(path.dirname(to));
  await fs.copyFile(from, to);
}

async function walkAndCopyMarkdown() {
  async function walk(dir) {
    const entries = await fs.readdir(dir, { withFileTypes: true });
    for (const ent of entries) {
      const full = path.join(dir, ent.name);
      if (ent.isDirectory()) {
        if (IGNORE_DIRS.has(ent.name)) continue;
        await walk(full);
      } else if (ent.isFile()) {
        if (ent.name === 'index.html' || ent.name === '404.html') continue;
        if (ent.name.endsWith('.md')) {
          // Copy markdown as-is; Starlight can infer titles from first heading
          const rel = path.relative(docsRoot, full);
          const dest = path.join(srcDir, rel);
          await ensureDir(path.dirname(dest));
          let txt = await fs.readFile(full, 'utf8');
          // Tabs: strip docsify markers; manual Tabs components can be added later if needed
          txt = txt.replace(/<!--\s*tabs:start\s*-->/g, '').replace(/<!--\s*tabs:end\s*-->/g, '');

          // Ensure frontmatter with a title for Starlight if missing
          const hasFrontmatter = /^---\n[\s\S]*?\n---\n/.test(txt);
          if (!hasFrontmatter) {
            // Try to extract first H1 as title
            const h1Match = txt.match(/^#\s+(.+)$/m);
            let inferred = h1Match ? h1Match[1].trim() : null;
            if (!inferred) {
              // Fallback to filename
              const base = path.basename(ent.name, path.extname(ent.name));
              inferred = base.replace(/[-_]+/g, ' ').replace(/\b\w/g, (m) => m.toUpperCase());
            }
            txt = `---\ntitle: ${inferred}\n---\n\n` + txt;
          }

          await fs.writeFile(dest, txt, 'utf8');
        }
      }
    }
  }
  await walk(docsRoot);
}

async function copySchemasAndMedia() {
  // schemas -> public/schemas
  const schemasSrc = path.join(docsRoot, 'schemas');
  try {
    const stat = await fs.stat(schemasSrc);
    if (stat.isDirectory()) {
      const dest = path.join(publicDir, 'schemas');
      await ensureDir(dest);
      async function walkSchemas(dir) {
        const entries = await fs.readdir(dir, { withFileTypes: true });
        for (const ent of entries) {
          const full = path.join(dir, ent.name);
          if (ent.isDirectory()) {
            await walkSchemas(full);
          } else if (ent.isFile()) {
            await copyFilePreserveDirs(full, path.join(publicDir));
          }
        }
      }
      await walkSchemas(schemasSrc);
    }
  } catch {}

  // _media -> public/_media
  const mediaSrc = path.join(docsRoot, '_media');
  try {
    const stat = await fs.stat(mediaSrc);
    if (stat.isDirectory()) {
      const dest = path.join(publicDir, '_media');
      await ensureDir(dest);
      async function walkMedia(dir) {
        const entries = await fs.readdir(dir, { withFileTypes: true });
        for (const ent of entries) {
          const full = path.join(dir, ent.name);
          if (ent.isDirectory()) {
            await walkMedia(full);
          } else if (ent.isFile()) {
            await copyFilePreserveDirs(full, path.join(publicDir));
          }
        }
      }
      await walkMedia(mediaSrc);
    }
  } catch {}
}

async function createHomeIndex() {
  // Ensure there is an index.md mirroring README.md for homepage
  const readmeSrc = path.join(docsRoot, 'README.md');
  try {
    let readme = await fs.readFile(readmeSrc, 'utf8');
    // Ensure frontmatter with title for home index
    const hasFrontmatter = /^---\n[\s\S]*?\n---\n/.test(readme);
    if (!hasFrontmatter) {
      const h1Match = readme.match(/^#\s+(.+)$/m);
      const inferred = h1Match ? h1Match[1].trim() : 'Home';
      readme = `---\ntitle: ${inferred}\ntemplate: doc\n---\n\n` + readme;
    } else {
      // Ensure the homepage uses the doc template so the sidebar is visible
      const hasTemplateDoc = /\ntemplate:\s*doc\n/.test(readme);
      if (!hasTemplateDoc) {
        readme = readme.replace(/^---\n/, '---\ntemplate: doc\n');
      }
    }
    const indexDest = path.join(srcDir, 'index.md');
    await ensureDir(path.dirname(indexDest));
    await fs.writeFile(indexDest, readme, 'utf8');
  } catch {}
}

async function main() {
  await ensureDir(srcDir);
  await ensureDir(publicDir);
  await Promise.all([
    walkAndCopyMarkdown(),
    copySchemasAndMedia(),
  ]);
  await createHomeIndex();
  console.log('✓ Content migration complete.');
  console.log('  - Markdown mirrored to src/content/docs');
  console.log('  - Schemas and media copied to public/');
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
