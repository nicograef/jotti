import { describe, expect, it } from 'vitest'

import {
  formatAlleAuswaehlenLabel,
  formatCents,
  formatEuro,
  formatEuroMitVorzeichen,
  formatPositionName,
  formatRelativeTime,
  parseCents,
} from './utils'

describe('formatCents', () => {
  it.each([
    [1250, '12,50'],
    [0, '0,00'],
    [5, '0,05'],
    [99, '0,99'],
    [500, '5,00'],
    [999999, '9999,99'],
    [-350, '-3,50'],
    [10, '0,10'],
  ])('formatCents(%i) = %s', (input, expected) => {
    expect(formatCents(input)).toBe(expected)
  })
})

describe('formatEuro', () => {
  it.each([
    [1250, '12,50\u00A0€'],
    [0, '0,00\u00A0€'],
    [-350, '-3,50\u00A0€'],
  ])('formatEuro(%i) = %s', (input, expected) => {
    expect(formatEuro(input)).toBe(expected)
  })

  it('trennt Betrag und € mit geschütztem Leerzeichen (U+00A0)', () => {
    expect(formatEuro(1250)).toBe('12,50\u00A0€')
    expect(formatEuro(1250)).not.toContain('12,50 €')
  })
})

describe('formatEuroMitVorzeichen', () => {
  it('setzt ein Plus nur bei positiven Beträgen', () => {
    expect(formatEuroMitVorzeichen(1250)).toBe(`+${formatEuro(1250)}`)
    expect(formatEuroMitVorzeichen(1)).toBe(`+${formatEuro(1)}`)
  })

  it('lässt Null ohne Vorzeichen', () => {
    expect(formatEuroMitVorzeichen(0)).toBe(formatEuro(0))
  })

  it('behält das Minus negativer Beträge', () => {
    expect(formatEuroMitVorzeichen(-350)).toBe(formatEuro(-350))
  })
})

describe('formatAlleAuswaehlenLabel', () => {
  it('nutzt „Alle …" mit Plural und „1 Position" im Singular (Standard-Variante)', () => {
    expect(formatAlleAuswaehlenLabel(3, 1250)).toBe(
      `Alle 3 Positionen auswählen · ${formatEuro(1250)}`,
    )
    expect(formatAlleAuswaehlenLabel(1, 700)).toBe(
      `1 Position auswählen · ${formatEuro(700)}`,
    )
  })

  it('nutzt „Meine …" in der Kassieren-Variante, Singular ohne Zahl', () => {
    expect(formatAlleAuswaehlenLabel(2, 1250, 'meine')).toBe(
      `Meine 2 Positionen auswählen · ${formatEuro(1250)}`,
    )
    expect(formatAlleAuswaehlenLabel(1, 700, 'meine')).toBe(
      `Meine Position auswählen · ${formatEuro(700)}`,
    )
  })
})

describe('formatPositionName', () => {
  it.each([
    ['Pommes', 'mit Ketchup', 'Pommes mit Ketchup'],
    ['Maß Bier', '', 'Maß Bier'],
    ['Cola', 'Cola', 'Cola Cola'],
  ])('formatPositionName(%s, %s) = %s', (produkt, variante, expected) => {
    expect(formatPositionName(produkt, variante)).toBe(expected)
  })
})

describe('formatRelativeTime', () => {
  // Fester Bezugspunkt für deterministische Grenzfälle.
  const now = new Date('2026-07-12T18:42:00')

  function vor(msVor: number): string {
    return new Date(now.getTime() - msVor).toISOString()
  }

  const MIN = 60_000
  const STD = 60 * MIN

  it('nennt Zeitpunkte unter einer Minute „gerade eben"', () => {
    expect(formatRelativeTime(vor(0), now)).toBe('gerade eben')
    expect(formatRelativeTime(vor(MIN - 1), now)).toBe('gerade eben')
  })

  it('zählt volle Minuten bis unter einer Stunde', () => {
    expect(formatRelativeTime(vor(MIN), now)).toBe('vor 1 min')
    expect(formatRelativeTime(vor(32 * MIN), now)).toBe('vor 32 min')
    expect(formatRelativeTime(vor(59 * MIN), now)).toBe('vor 59 min')
  })

  it('zählt volle Stunden bis unter sechs Stunden', () => {
    expect(formatRelativeTime(vor(STD), now)).toBe('vor 1 Std')
    expect(formatRelativeTime(vor(5 * STD + 59 * MIN), now)).toBe('vor 5 Std')
  })

  it('zeigt ab sechs Stunden die absolute Uhrzeit desselben Tages', () => {
    expect(formatRelativeTime('2026-07-12T09:05:00', now)).toBe('09:05')
    expect(formatRelativeTime('2026-07-12T00:00:00', now)).toBe('00:00')
  })

  it('ergänzt an früheren Tagen das Datum', () => {
    expect(formatRelativeTime('2026-07-11T18:42:00', now)).toBe('11.7., 18:42')
    expect(formatRelativeTime('2025-12-24T20:00:00', now)).toBe('24.12., 20:00')
  })
})

describe('parseCents', () => {
  it.each([
    ['12,50', 1250],
    ['12.50', 1250],
    ['0,05', 5],
    ['0', 0],
    ['-3,50', -350],
    ['', 0],
    ['abc', 0],
    ['10,', 1000],
    [',5', 50],
    ['12,505', 0],
    ['1,2,3', 0],
    ['12,50 €', 0],
  ])('parseCents(%s) = %i', (input, expected) => {
    expect(parseCents(input)).toBe(expected)
  })
})
