import { dirname, join } from 'node:path/posix'

// In `docs/` stehen repo-relative Markdown-Links (`[x.md](x.md)`, auch mit
// `#anker`), damit die Vorschau auf GitHub und im Editor gültig bleibt.

export interface RewriteDocLinkOptions {
  /** Link-Ziel wie im Markdown geschrieben, z. B. `compliance.md#anker`, `../TERMS.md`. */
  target: string
  /** Pfad des Quelldokuments relativ zu `docs/`, z. B. `verfahrensdokumentation.md`. */
  sourcePath: string
  /** Veröffentlichte Dokumente als Pfade relativ zu `docs/`. */
  publishedDocs: readonly string[]
  /** Basis-URL für GitHub-`blob`-Links, z. B. `https://github.com/nicograef/jotti/blob/main`. */
  repoBaseUrl: string
}

export function rewriteDocLink({
  target,
  sourcePath,
  publishedDocs,
  repoBaseUrl,
}: RewriteDocLinkOptions): string {
  if (/^(https?:|mailto:)/i.test(target)) return target

  const hashIndex = target.indexOf('#')
  const path = hashIndex === -1 ? target : target.slice(0, hashIndex)
  const anchor = hashIndex === -1 ? '' : target.slice(hashIndex)

  if (path === '') return target

  // `join` normalisiert `.`/`..`, sodass `../TERMS.md` aus `docs/` zu `TERMS.md`
  // außerhalb von `docs/` wird.
  const resolved = join(dirname(join('docs', sourcePath)), path)

  const docsRelative = resolved.startsWith('docs/')
    ? resolved.slice('docs/'.length)
    : null

  if (docsRelative !== null && publishedDocs.includes(docsRelative)) {
    const slug = docsRelative.replace(/\.[^./]+$/, '')
    return `/docs/${slug}/${anchor}`
  }

  // Privates Dokument oder Datei außerhalb `docs/`: auf die GitHub-Quelle zeigen.
  return `${repoBaseUrl}/${resolved}${anchor}`
}
