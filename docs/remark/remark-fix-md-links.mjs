// Remark plugin to rewrite local .md links to pretty URLs for Starlight
// Examples:
//   installation.md        -> installation/
//   guide/README.md        -> guide/README/
//   ../types/config.md#x   -> ../types/config/#x
// External links (http/https/mailto) and absolute URLs with file extensions are left alone.

export default function remarkFixMdLinks() {
  return function (_, file) {
    const visit = (node, fn) => {
      if (!node || typeof node !== 'object') return;
      if (Array.isArray(node.children)) {
        for (const child of node.children) visit(child, fn);
      }
      fn(node);
    };

    visit(file.data.astro?.frontmatter ? file.data.astro.content : file, (node) => {
      if (!node || node.type !== 'link' || !node.url) return;
      const url = String(node.url);
      if (/^(https?:|mailto:|tel:|#)/.test(url)) return;
      // Ignore links with extensions other than .md
      const match = url.match(/^(.*?)(\.md)(#.*)?$/);
      if (!match) return;
      const base = match[1];
      const hash = match[3] || '';
      // Keep relative/absolute pathing; replace .md with trailing slash
      let pretty = base;
      // Ensure a trailing slash if not already present
      if (!pretty.endsWith('/')) pretty += '/';
      node.url = pretty + (hash ? hash : '');
    });
  };
}
