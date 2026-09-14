import { relative } from 'node:path'

import { rewriteDocLink } from './link-rewriter'
import { publishedDocs } from './published-docs'

// mdast-Adapter für `rewriteDocLink`: stellt Quellpfad und Link-Knoten bei.

export interface RemarkDocLinksOptions {
  /** Absoluter Pfad zum top-level `docs/`-Verzeichnis. */
  docsDir: string
  /** Basis-URL für GitHub-`blob`-Links, z. B. `https://github.com/nicograef/jotti/blob/main`. */
  repoBaseUrl: string
}

interface MdastNode {
  type: string
  url?: string
  children?: MdastNode[]
}

interface VFileLike {
  path?: string
  history?: string[]
}

export function remarkDocLinks({
  docsDir,
  repoBaseUrl,
}: RemarkDocLinksOptions) {
  return function transform(tree: MdastNode, file: VFileLike): void {
    const filePath = file.path ?? file.history?.at(-1)
    if (!filePath) return

    const sourcePath = relative(docsDir, filePath).split('\\').join('/')
    // Nur Dateien innerhalb von docs/ verarbeiten.
    if (sourcePath.startsWith('..')) return

    visitLinks(tree, (node) => {
      if (node.url === undefined) return
      node.url = rewriteDocLink({
        target: node.url,
        sourcePath,
        publishedDocs,
        repoBaseUrl,
      })
    })
  }
}

// Eigener Walk statt `unist-util-visit`: spart eine Abhängigkeit.
function visitLinks(node: MdastNode, fn: (node: MdastNode) => void): void {
  if (node.type === 'link' || node.type === 'definition') fn(node)
  for (const child of node.children ?? []) visitLinks(child, fn)
}
