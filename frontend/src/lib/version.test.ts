import { describe, expect, it } from 'vitest'

import { CLIENT_VERSION, istVersionsabweichung } from './version'

describe('CLIENT_VERSION', () => {
  // Gepinnt: Der Vergleich schlägt in Dev, E2E und Tests nie an, weil auf
  // beiden Seiten der `define`-Default `dev` steht.
  it('ist ohne Build-Arg der Default dev', () => {
    expect(CLIENT_VERSION).toBe('dev')
  })
})

describe('istVersionsabweichung', () => {
  // Nur zwei verschiedene echte Releases ergeben eine Abweichung; jede Zeile
  // mit `dev` steht für Dev, E2E oder Test.
  const faelle: [string, string, boolean][] = [
    ['dev', 'dev', false],
    ['dev', 'v1.2.3', false],
    ['v1.2.3', 'dev', false],
    ['v1.2.3', 'v1.2.3', false],
    ['v1.2.3', 'v1.2.4', true],
    ['dev-a1b2c3d', 'v1.2.3', false],
    ['v1.2.3-rc1', 'v1.2.3', true],
  ]

  it.each(faelle)(
    'Client %s gegen Server %s ergibt %s',
    (clientVersion, serverVersion, erwartet) => {
      expect(istVersionsabweichung(clientVersion, serverVersion)).toBe(erwartet)
    },
  )
})
