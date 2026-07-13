import { describe, expect, it } from 'vitest'

import {
  beschreibeFehlBons,
  DruckstationConfigSchema,
  formatDruckauftragReferenz,
  formatDruckfehler,
  hatBonmodus,
  validateDruckerIp,
} from './DruckstationBackend'

describe('validateDruckerIp', () => {
  it('accepts an empty IP (no printer)', () => {
    expect(validateDruckerIp('')).toBeNull()
  })

  it('accepts a valid IPv4 address', () => {
    expect(validateDruckerIp('192.168.1.50')).toBeNull()
  })

  it('rejects an invalid IPv4 address with a field message', () => {
    expect(validateDruckerIp('999.999.999.999')).toBe('Ungültige IPv4-Adresse')
  })

  it('rejects a non-IP string', () => {
    expect(validateDruckerIp('not-an-ip')).toBe('Ungültige IPv4-Adresse')
  })
})

describe('hatBonmodus', () => {
  it('is true for product categories and abholbon', () => {
    expect(hatBonmodus('essen')).toBe(true)
    expect(hatBonmodus('getraenk')).toBe(true)
    expect(hatBonmodus('sonstiges')).toBe(true)
    expect(hatBonmodus('abholbon')).toBe(true)
  })

  it('is false for kassenbeleg', () => {
    expect(hatBonmodus('kassenbeleg')).toBe(false)
  })
})

describe('formatDruckauftragReferenz', () => {
  it.each([
    ['bestellung-aufgenommen:86', 'Bestellung Nr. 86'],
    ['zahlung-kassiert:12', 'Zahlung Nr. 12'],
    ['direktverkauf-getaetigt:3', 'Direktverkauf Nr. 3'],
    ['direktverkauf-storniert:7', 'Direktverkauf-Storno Nr. 7'],
    ['stornierung-erteilt:41', 'Stornierung Nr. 41'],
    ['testdruck:essen', 'Testbon Essen'],
    ['testdruck:getraenk', 'Testbon Getränk'],
    ['testdruck:kassenbeleg', 'Testbon Kassenbeleg'],
  ])('übersetzt %s zu "%s"', (referenz, erwartet) => {
    expect(formatDruckauftragReferenz(referenz)).toBe(erwartet)
  })

  it('zeigt einen unbekannten Testbon-Rest wörtlich', () => {
    expect(formatDruckauftragReferenz('testdruck:foo')).toBe('Testbon foo')
  })

  it('fällt bei unbekanntem Format auf den Rohwert zurück', () => {
    expect(formatDruckauftragReferenz('unbekanntes-format:99')).toBe(
      'unbekanntes-format:99',
    )
  })

  it('fällt bei fehlendem Trenner auf den Rohwert zurück', () => {
    expect(formatDruckauftragReferenz('ohne-doppelpunkt')).toBe(
      'ohne-doppelpunkt',
    )
  })
})

describe('beschreibeFehlBons', () => {
  it('nennt reine Arbeitsbons „Bon" und ordnet sie der Küche zu', () => {
    expect(beschreibeFehlBons(['arbeitsbon', 'arbeitsbon'])).toEqual({
      singular: 'Bon',
      plural: 'Bons',
      kuecheBetroffen: true,
    })
  })

  it('nennt Kassenbelege beim Namen und behauptet kein Küchenproblem', () => {
    expect(beschreibeFehlBons(['kassenbeleg'])).toEqual({
      singular: 'Kassenbeleg',
      plural: 'Kassenbelege',
      kuecheBetroffen: false,
    })
  })

  it('nennt Testbons neutral ohne Küchenbezug', () => {
    expect(beschreibeFehlBons(['testbon'])).toEqual({
      singular: 'Testbon',
      plural: 'Testbons',
      kuecheBetroffen: false,
    })
  })

  it('fällt bei gemischten Bon-Arten auf den neutralen Oberbegriff zurück', () => {
    expect(beschreibeFehlBons(['arbeitsbon', 'kassenbeleg'])).toEqual({
      singular: 'Bon',
      plural: 'Bons',
      kuecheBetroffen: false,
    })
  })

  it('fällt bei unbekannter Bon-Art auf den neutralen Oberbegriff zurück', () => {
    expect(beschreibeFehlBons(['unbekannt'])).toEqual({
      singular: 'Bon',
      plural: 'Bons',
      kuecheBetroffen: false,
    })
  })
})

describe('formatDruckfehler', () => {
  it.each([
    [
      'drucker 192.168.1.5: nicht erreichbar: dial tcp',
      'Drucker nicht erreichbar',
    ],
    ['papier leer (status=0x08)', 'Papier leer'],
    ['Drucker meldet: Abdeckung offen', 'Abdeckung offen'],
    [
      'drucker 192.168.1.5: senden fehlgeschlagen: EOF',
      'Übertragung fehlgeschlagen',
    ],
  ])('übersetzt „%s" zu „%s"', (roh, erwartet) => {
    expect(formatDruckfehler(roh)).toBe(erwartet)
  })

  it('fällt bei unbekanntem Jargon auf eine neutrale Sammelmeldung zurück', () => {
    expect(formatDruckfehler('ungueltiges Base64: illegal data')).toBe(
      'Druckfehler',
    )
  })
})

describe('DruckstationConfigSchema', () => {
  it('accepts a product station with bonmodus', () => {
    const parsed = DruckstationConfigSchema.parse({
      kategorie: 'essen',
      druckerIp: '192.168.1.51',
      bonmodus: 'pro_position',
    })
    expect(parsed.bonmodus).toBe('pro_position')
  })

  it('accepts a kassenbeleg station with empty bonmodus', () => {
    const parsed = DruckstationConfigSchema.parse({
      kategorie: 'kassenbeleg',
      druckerIp: '192.168.1.60',
      bonmodus: '',
    })
    expect(parsed.kategorie).toBe('kassenbeleg')
    expect(parsed.bonmodus).toBe('')
  })
})
