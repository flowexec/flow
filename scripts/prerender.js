const fs = require('fs');
const path = require('path');
const Renderer = require('docsify-server-renderer');

// Recursively find all `_sidebar.md` files under ./docs
function* walk(dir) {
  const items = fs.readdirSync(dir, { withFileTypes: true });
  for (const item of items) {
    const full = path.join(dir, item.name);
    if (item.isDirectory()) {
      yield* walk(full);
    } else if (item.isFile()) {
      yield full;
    }
  }
}

function discoverLinks() {
  const docsRoot = path.resolve('./docs');
  const sidebars = Array.from(walk('./docs')).filter(p => path.basename(p) === '_sidebar.md');
  const links = [];
  for (const sb of sidebars) {
    const md = fs.readFileSync(sb, 'utf-8');
    const baseDirRel = path.posix.dirname(
      path.posix.relative(docsRoot, path.resolve(path.dirname(sb)))
    );
    const matches = Array.from(md.matchAll(/\[[^\]]+\]\(([^)]+)\)/g))
      .map(m => m[1])
      .filter(href => href && !href.startsWith('http'))
      .map(href => {
        // strip optional Markdown link title and angle brackets, drop anchors/queries
        let h = href.replace(/\s+["'][^"')]*["']\s*$/, '').trim();
        h = h.replace(/^<(.+)>$/, '$1');
        h = h.split('#')[0].split('?')[0];
        // if already absolute from docs root
        if (h.startsWith('/')) {
          return path.posix.normalize(h);
        }
        // resolve relative to the sidebar's directory, then make it absolute from docs root
        const joined = path.posix.join('/', baseDirRel, h);
        return path.posix.normalize(joined);
      })
      // skip attempts to traverse above root
      .filter(h => !h.startsWith('/..'));
    links.push(...matches);
  }
  return links;
}

// Normalize to routes (remove .md, ensure leading slash)
function toRoutes(links) {
  const set = new Set(['/']);
  for (let href of links) {
    // strip optional Markdown link title in quotes: (/path "Title") or (/path 'Title')
    href = href.replace(/\s+["'][^"')]*["']\s*$/, '').trim();
    // remove surrounding angle brackets if present: (<path>)
    href = href.replace(/^<(.+)>$/, '$1');
    // drop anchors and query
    href = href.split('#')[0].split('?')[0];
    // remove trailing 'index' or 'README' if present
    href = href.replace(/(README|readme|index)\.md$/, '');
    // remove .md extension
    href = href.replace(/\.md$/, '');
    if (!href) href = '/';
    if (!href.startsWith('/')) href = '/' + href;
    // trim trailing slash, but keep root '/'
    if (href.length > 1) href = href.replace(/\/$/, '');
    set.add(href);
  }
  return Array.from(set);
}

const links = discoverLinks();
const routes = toRoutes(links);
console.log('Discovered routes:', routes);

// Prepare renderer with the same config as docs/index.html,
// and basePath pointing to the docs directory so local files resolve.
const template = fs.readFileSync('./docs/index.html', 'utf-8');
const renderer = new Renderer({
  template,
  config: {
    name: 'flow',
    repo: 'flowexec/flow',
    homepage: 'README.md',
    // Disable relativePath in SSR to avoid undefined currentRoute during path parsing
    relativePath: false,
    themeColor: '#D699B6',
    loadSidebar: true,
    formatUpdated: '{MM}/{DD} {HH}:{mm}',
    maxLevel: 4,
    subMaxLevel: 3,
    coverpage: true,
    onlyCover: false,
    nameLink: 'README',
    auto2top: true,
    routerMode: 'history',
    alias: {
      '/releases/(.*)': 'https://github.com/flowexec/flow/releases/$1',
      '/schemas/(.*)': 'https://github.com/flowexec/flow/tree/main/schemas/$1',
      '/issues/(.*)': 'https://github.com/flowexec/flow/issues/$1',
      '/(.*)/_sidebar.md': '$1/_sidebar.md',
    },
    basePath: path.resolve('./docs')
  }
});

// Work around SSR router.parse(undefined) by defaulting to current URL
if (renderer && renderer.router && typeof renderer.router.parse === 'function') {
  const _parse = renderer.router.parse.bind(renderer.router);
  renderer.router.parse = function (p) {
    return _parse(p || renderer.url || '/');
  };
}

function rimraf(target) {
  if (fs.existsSync(target)) {
    fs.rmSync(target, { recursive: true, force: true });
  }
}

function copyDir(src, dest) {
  const stat = fs.statSync(src);
  if (stat.isDirectory()) {
    fs.mkdirSync(dest, { recursive: true });
    for (const entry of fs.readdirSync(src)) {
      const s = path.join(src, entry);
      const d = path.join(dest, entry);
      copyDir(s, d);
    }
  } else if (stat.isFile()) {
    fs.mkdirSync(path.dirname(dest), { recursive: true });
    fs.copyFileSync(src, dest);
  }
}

async function build() {
  // Clean dist and copy all docs assets first (so client-side nav still works)
  rimraf('./dist');
  copyDir('./docs', './dist');

  for (const route of routes) {
    const html = await renderer.renderToString(route);
    const filePath =
      route === '/'
        ? './dist/index.html'
        : path.join('./dist', route, 'index.html');

    fs.mkdirSync(path.dirname(filePath), { recursive: true });
    fs.writeFileSync(filePath, html, 'utf-8');
    console.log(`Rendered ${route} -> ${filePath}`);
  }

  // Optional: create a 404.html fallback (serve homepage)
  try {
    const indexHtml = fs.readFileSync('./dist/index.html', 'utf-8');
    fs.writeFileSync('./dist/404.html', indexHtml, 'utf-8');
  } catch {}
}

build().catch(err => {
  console.error(err);
  process.exit(1);
});
