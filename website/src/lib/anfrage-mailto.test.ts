import { describe, expect, it } from 'vitest'

import type { AnfrageFelder } from './anfrage-mailto'
import { buildMailtoUrl, hatFehler, validateAnfrage } from './anfrage-mailto'
import { betreiberEmail } from './links'

// Vollständig ausgefüllte Basis; einzelne Felder werden je Fall überschrieben.
function felder(overrides: Partial<AnfrageFelder> = {}): AnfrageFelder {
  return {
    verein: 'TSV Musterhausen e.V.',
    name: 'Erika Mustermann',
    email: 'vorstand@musterhausen.de',
    art: 'Eingetragener Verein (e.V.)',
    message: '',
    ...overrides,
  }
}

describe('validateAnfrage', () => {
  it('meldet keine Fehler bei vollständiger Eingabe', () => {
    const fehler = validateAnfrage(felder())
    expect(fehler).toEqual({})
    expect(hatFehler(fehler)).toBe(false)
  })

  it('meldet jedes leere Pflichtfeld (verein, name, email)', () => {
    const fehler = validateAnfrage(
      felder({ verein: '', name: '', email: '' }),
    )
    expect(fehler.verein).toBeTruthy()
    expect(fehler.name).toBeTruthy()
    expect(fehler.email).toBeTruthy()
    expect(hatFehler(fehler)).toBe(true)
  })

  it('wertet reine Leerzeichen als leer', () => {
    const fehler = validateAnfrage(felder({ verein: '   ', name: '\t' }))
    expect(fehler.verein).toBeTruthy()
    expect(fehler.name).toBeTruthy()
  })

  it('lehnt eine E-Mail ohne gültiges Format ab', () => {
    expect(validateAnfrage(felder({ email: 'keine-mail' })).email).toBeTruthy()
    expect(validateAnfrage(felder({ email: 'a@b' })).email).toBeTruthy()
  })

  it('akzeptiert eine gültige E-Mail und validiert art/message nicht', () => {
    const fehler = validateAnfrage(
      felder({ email: 'a.b+tag@verein-example.de', art: '', message: '' }),
    )
    expect(fehler).toEqual({})
  })
})

describe('buildMailtoUrl', () => {
  it('setzt die Betreiber-Adresse als Empfänger', () => {
    const url = new URL(buildMailtoUrl(felder()))
    expect(url.protocol).toBe('mailto:')
    expect(url.pathname).toBe(betreiberEmail)
  })

  it('legt Betreff und alle Feldwerte korrekt in der URL ab', () => {
    const params = new URLSearchParams(
      new URL(
        buildMailtoUrl(
          felder({ message: 'Für unser Sommerfest im Juli.' }),
        ),
      ).search,
    )
    expect(params.get('subject')).toBe(
      'Nutzungsvereinbarung anfragen – TSV Musterhausen e.V.',
    )
    const body = params.get('body') ?? ''
    expect(body).toContain('Verein / Organisation: TSV Musterhausen e.V.')
    expect(body).toContain('Ansprechpartner:in: Erika Mustermann')
    expect(body).toContain('E-Mail: vorstand@musterhausen.de')
    expect(body).toContain('Rechtsform: Eingetragener Verein (e.V.)')
    expect(body).toContain('Nachricht:\nFür unser Sommerfest im Juli.')
  })

  it('lässt den Nachrichten-Block weg, wenn keine Nachricht eingegeben wurde', () => {
    const body =
      new URLSearchParams(new URL(buildMailtoUrl(felder())).search).get(
        'body',
      ) ?? ''
    expect(body).not.toContain('Nachricht:')
  })

  it('trimmt die Feldwerte für einen sauberen Entwurf', () => {
    const params = new URLSearchParams(
      new URL(
        buildMailtoUrl(
          felder({ verein: '  TSV Musterhausen e.V.  ', message: '  Hallo  ' }),
        ),
      ).search,
    )
    expect(params.get('subject')).toBe(
      'Nutzungsvereinbarung anfragen – TSV Musterhausen e.V.',
    )
    expect(params.get('body')).toContain('Nachricht:\nHallo')
  })

  it('encodiert Umlaute prozentual (äöüß)', () => {
    const url = buildMailtoUrl(felder({ verein: 'Schützenverein Grünäöüß' }))
    // ä ö ü ß dürfen nicht roh in der URL stehen, sondern als %C3%…
    expect(url).not.toContain('Grünäöüß')
    expect(url).toContain(encodeURIComponent('Schützenverein Grünäöüß'))
    // Round-trip: decodiert steht der Originaltext wieder im Betreff.
    const subject = new URLSearchParams(new URL(url).search).get('subject')
    expect(subject).toBe('Nutzungsvereinbarung anfragen – Schützenverein Grünäöüß')
  })

  it('encodiert Zeilenumbrüche im Body als %0A', () => {
    const url = buildMailtoUrl(felder({ message: 'Zeile 1\nZeile 2' }))
    expect(url).toContain('%0A')
    const body = new URLSearchParams(new URL(url).search).get('body') ?? ''
    expect(body).toContain('Zeile 1\nZeile 2')
  })

  it('encodiert Sonderzeichen (& ? = + und Leerzeichen)', () => {
    const url = buildMailtoUrl(felder({ message: 'a & b ? c = d + e' }))
    // Kein rohes Sonderzeichen im Query-Teil außer den Trennern der mailto-URL.
    expect(url).toContain('%26') // &
    expect(url).toContain('%3F') // ?
    expect(url).toContain('%3D') // =
    expect(url).toContain('%2B') // +
    expect(url).toContain('%20') // Leerzeichen
    const body = new URLSearchParams(new URL(url).search).get('body') ?? ''
    expect(body).toContain('a & b ? c = d + e')
  })
})
