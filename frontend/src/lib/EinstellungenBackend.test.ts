import { describe, expect, it } from 'vitest'

import { BondruckEinstellungenSchema } from './EinstellungenBackend'

describe('BondruckEinstellungenSchema', () => {
  it('accepts valid bondruck settings with abholbon mode', () => {
    const parsed = BondruckEinstellungenSchema.parse({
      kassenbelegDruckerIp: '192.168.1.80',
      direktverkaufModus: 'abholbon',
      abholbonDruckerIp: '192.168.1.81',
    })

    expect(parsed).toEqual({
      kassenbelegDruckerIp: '192.168.1.80',
      direktverkaufModus: 'abholbon',
      abholbonDruckerIp: '192.168.1.81',
    })
  })

  it('accepts empty IPs with kein_bon mode', () => {
    const parsed = BondruckEinstellungenSchema.parse({
      kassenbelegDruckerIp: '',
      direktverkaufModus: 'kein_bon',
      abholbonDruckerIp: '',
    })

    expect(parsed.direktverkaufModus).toBe('kein_bon')
  })

  it('rejects unknown direktverkauf mode', () => {
    expect(() =>
      BondruckEinstellungenSchema.parse({
        kassenbelegDruckerIp: '192.168.1.80',
        direktverkaufModus: 'invalid_mode',
        abholbonDruckerIp: '192.168.1.81',
      }),
    ).toThrow()
  })

  it('rejects invalid abholbon IPv4', () => {
    expect(() =>
      BondruckEinstellungenSchema.parse({
        kassenbelegDruckerIp: '192.168.1.80',
        direktverkaufModus: 'abholbon',
        abholbonDruckerIp: '999.999.999.999',
      }),
    ).toThrow()
  })
})
