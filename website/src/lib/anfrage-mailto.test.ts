import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import type { AnfrageFelder } from './anfrage-mailto'
import {
  buildAnfrageMail,
  buildMailtoUrl,
  hatFehler,
  validateAnfrage,
} from './anfrage-mailto'
import { betreiberEmail } from './links'

function felder(overrides: Partial<AnfrageFelder> = {}): AnfrageFelder {
  return {
    verein: 'TSV Musterhausen e.V.',
    sitz: 'Musterhausen',
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

  it('meldet jedes leere Pflichtfeld (verein, sitz, name, email)', () => {
    const fehler = validateAnfrage(
      felder({ verein: '', sitz: '', name: '', email: '' }),
    )
    expect(fehler.verein).toBeTruthy()
    expect(fehler.sitz).toBeTruthy()
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

describe('buildAnfrageMail', () => {
  it('liefert Empfänger, Betreff und Text deckungsgleich mit der mailto-URL', () => {
    const eingabe = felder({ message: 'Grüße & Fragen: 2 + 2?\nZeile 2' })
    const mail = buildAnfrageMail(eingabe)
    const raw = buildMailtoUrl(eingabe)
    const url = new URL(raw)

    // RFC 6068: ein "+" ist im mailto-Body wörtlich, Leerzeichen müssen %20 sein.
    expect(raw).not.toContain('+')
    expect(url.pathname).toBe(mail.empfaenger)
    expect(url.searchParams.get('subject')).toBe(mail.betreff)
    expect(url.searchParams.get('body')).toBe(mail.text)
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
        buildMailtoUrl(felder({ message: 'Für unser Sommerfest im Juli.' })),
      ).search,
    )
    expect(params.get('subject')).toBe(
      'Nutzungsvereinbarung jotti — TSV Musterhausen e.V.',
    )
    const body = params.get('body') ?? ''
    expect(body).toContain('wir sind TSV Musterhausen e.V.')
    expect(body).toContain('Rechtsform: Eingetragener Verein (e.V.)')
    expect(body).toContain('Sitz: Musterhausen')
    expect(body).toContain(
      'Ansprechperson: Erika Mustermann, vorstand@musterhausen.de',
    )
    expect(body).toContain('Nachricht:\nFür unser Sommerfest im Juli.')
  })

  it('nennt im Annahmesatz die Fassung aus dem Stand von TERMS.md', () => {
    const terms = readFileSync(
      new URL('../../../TERMS.md', import.meta.url),
      'utf8',
    )
    const stand = /^\*\*Stand: (.+)\*\*$/m.exec(terms)?.[1]
    const vorlage = /in der Fassung vom (.+)$/m.exec(terms)?.[1]

    expect(stand).toBeTruthy()
    expect(vorlage).toBe(stand)
    expect(buildAnfrageMail(felder()).text).toContain(
      `in der Fassung vom ${stand ?? ''} (https://github.com/nicograef/jotti/blob/main/TERMS.md).`,
    )
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
      'Nutzungsvereinbarung jotti — TSV Musterhausen e.V.',
    )
    expect(params.get('body')).toContain('Nachricht:\nHallo')
  })
})
