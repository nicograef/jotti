import { dirname, join } from 'node:path/posix'

// Reine Kernlogik der Querverweis-Auflösung (PRD: „remark-Link-Rewriter").
//
// Autoren und Agenten schreiben in `docs/` weiter normale repo-relative
// Markdown-Links (`[x.md](x.md)`, auch mit `#anker`), damit die Vorschau auf
// GitHub und im Editor gültig bleibt. Aus einem solchen Link-Ziel berechnet
// diese Funktion das passende `href` für die gebaute Website.

export interface RewriteDocLinkOptions {
  /** Link-Ziel wie im Markdown geschrieben, z. B. `compliance.md#anker`, `../TERMS.md`. */
  target: string
  /** Pfad des Quelldokuments relativ zu `docs/`, z. B. `verfahrensdokumentation.md`. */
  sourcePath: string
  /** Veröffentlichte Dokumente als Pfade relativ zu `docs/` (Veröffentlichungs-Auswahl). */
  publishedDocs: readonly string[]
  /** Basis-URL für GitHub-`blob`-Links, z. B. `https://github.com/nicograef/jotti/blob/main`. */
  repoBaseUrl: string
}

/**
 * Bildet ein repo-relatives Markdown-Link-Ziel auf das Website-`href` ab:
 *
 * - Ziel ist ein veröffentlichtes Dokument → Website-Route `/docs/<slug>/`
 *   (Anker als Slug der Zielüberschrift bleibt erhalten).
 * - Ziel ist ein privates Dokument oder eine Datei außerhalb `docs/` →
 *   absolute GitHub-`blob`-URL auf die Quelle.
 * - Externe Links (`http`, `https`, `mailto`) und reine Anker (`#…`) bleiben
 *   unverändert.
 */
export function rewriteDocLink({
  target,
  sourcePath,
  publishedDocs,
  repoBaseUrl,
}: RewriteDocLinkOptions): string {
  // Externe Links und mailto bleiben unverändert.
  if (/^(https?:|mailto:)/i.test(target)) return target

  const hashIndex = target.indexOf('#')
  const path = hashIndex === -1 ? target : target.slice(0, hashIndex)
  const anchor = hashIndex === -1 ? '' : target.slice(hashIndex)

  // Reiner Anker ohne Pfad zeigt auf eine Überschrift im selben Dokument.
  if (path === '') return target

  // Ziel relativ zum Quelldokument auf einen repo-root-relativen Pfad auflösen.
  // `join` normalisiert dabei `.`/`..`, sodass `../TERMS.md` aus `docs/` zu
  // `TERMS.md` (außerhalb von `docs/`) wird.
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
