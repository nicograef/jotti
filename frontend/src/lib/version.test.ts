import { describe, expect, it } from 'vitest'

import { CLIENT_VERSION, istVersionsabweichung } from './version'

describe('CLIENT_VERSION', () => {
  // The guard tests only prove something while the test build is a real release.
  it('ist im Test eine echte Release-Version, damit der Vergleich scharf ist', () => {
    expect(istVersionsabweichung(CLIENT_VERSION, 'v0.0.1')).toBe(true)
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
