import { describe, expect, it } from 'vitest'

import { beschreibeFehlBons, validateDruckerIp } from './DruckstationBackend'

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
