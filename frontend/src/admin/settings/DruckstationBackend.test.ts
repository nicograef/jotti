import { describe, expect, it } from 'vitest'

import {
  DruckstationConfigSchema,
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
