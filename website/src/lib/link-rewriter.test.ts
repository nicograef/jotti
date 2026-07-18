import { describe, expect, it } from 'vitest'

import { rewriteDocLink } from './link-rewriter'
import { publishedDocs } from './published-docs'

const repoBaseUrl = 'https://github.com/nicograef/jotti/blob/main'

// Kleiner Wrapper: feste Map und Repo-Basis, damit die Fälle knapp bleiben.
function rewrite(target: string, sourcePath = 'verfahrensdokumentation.md') {
  return rewriteDocLink({ target, sourcePath, publishedDocs, repoBaseUrl })
}

describe('rewriteDocLink', () => {
  it('mappt ein veröffentlichtes Dokument auf seine Website-Route', () => {
    expect(rewrite('compliance.md')).toBe('/docs/compliance/')
  })

  it('hängt den Anker an die Website-Route an (Slug der Zielüberschrift)', () => {
    expect(rewrite('compliance.md#42-anforderungen')).toBe(
      '/docs/compliance/#42-anforderungen',
    )
  })

  it('löst Links aus einem Unterordner über `../` auf', () => {
    const href = rewriteDocLink({
      target: '../compliance.md#3-tse',
      sourcePath: 'leitfaden/installation.md',
      publishedDocs,
      repoBaseUrl,
    })
    expect(href).toBe('/docs/compliance/#3-tse')
  })

  it('erzeugt für veröffentlichte Dokumente in Unterordnern verschachtelte Routen', () => {
    const href = rewriteDocLink({
      target: 'checkliste.md',
      sourcePath: 'leitfaden/installation.md',
      publishedDocs: ['leitfaden/checkliste.md'],
      repoBaseUrl,
    })
    expect(href).toBe('/docs/leitfaden/checkliste/')
  })

  it('schreibt einen repo-relativen Link auf die Betriebsarten-Seite um', () => {
    const href = rewriteDocLink({
      target: 'betriebsarten.md',
      sourcePath: 'leitfaden/was-ist-jotti.md',
      publishedDocs,
      repoBaseUrl,
    })
    expect(href).toBe('/docs/leitfaden/betriebsarten/')
  })

  it('verweist auf private Dokumente per GitHub-URL', () => {
    expect(rewrite('handbuch.md#313-tse-architektur')).toBe(
      `${repoBaseUrl}/docs/handbuch.md#313-tse-architektur`,
    )
  })

  it('verweist auf Dateien außerhalb von docs/ per GitHub-URL', () => {
    expect(rewrite('../TERMS.md', 'lizenzmodell.md')).toBe(
      `${repoBaseUrl}/TERMS.md`,
    )
    expect(rewrite('../LICENSE', 'lizenzmodell.md')).toBe(
      `${repoBaseUrl}/LICENSE`,
    )
  })

  it('lässt externe Links und mailto unverändert', () => {
    expect(rewrite('https://github.com/nicograef/jotti')).toBe(
      'https://github.com/nicograef/jotti',
    )
    expect(rewrite('mailto:graef.nico@gmail.com')).toBe(
      'mailto:graef.nico@gmail.com',
    )
  })

  it('lässt reine Anker (gleiches Dokument) unverändert', () => {
    expect(rewrite('#36-das-festzelt-muster')).toBe('#36-das-festzelt-muster')
  })
})
