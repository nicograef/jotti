import { beforeEach, describe, expect, it } from 'vitest'

import { getArbeitsmodus } from './lib/arbeitsmodus'
import {
  ServiceDirektverkaufLoader,
  ServiceIndexRedirect,
  ServiceTableGuard,
  ServiceTischauswahlLoader,
} from './routes'

function locationOf(result: unknown): string {
  expect(result).toBeInstanceOf(Response)
  if (!(result instanceof Response)) {
    throw new Error('Expected redirect response')
  }
  expect(result.status).toBe(302)
  return result.headers.get('Location') ?? ''
}

describe('ServiceTableGuard', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('redirects to table selection for invalid tischId', () => {
    const result = ServiceTableGuard({
      params: { tischId: 'abc' },
    } as never)

    expect(locationOf(result)).toBe('/service/tische')
  })

  it('allows valid positive integer tischId and persists Tischservice', () => {
    const result = ServiceTableGuard({
      params: { tischId: '12' },
    } as never)

    expect(result).toBeUndefined()
    expect(getArbeitsmodus()).toBe('tischservice')
  })
})

describe('ServiceIndexRedirect', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('leitet ohne Präferenz auf die Tischauswahl', () => {
    expect(locationOf(ServiceIndexRedirect())).toBe('tische')
  })

  it('leitet bei gespeichertem Direktverkauf auf den Direktverkauf', () => {
    ServiceDirektverkaufLoader()

    expect(locationOf(ServiceIndexRedirect())).toBe('direktverkauf')
  })

  it('leitet bei gespeichertem Tischservice auf die Tischauswahl', () => {
    ServiceTischauswahlLoader()

    expect(locationOf(ServiceIndexRedirect())).toBe('tische')
  })
})

describe('Modus-Route-Loader', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('Tischauswahl-Loader persistiert Tischservice', () => {
    ServiceTischauswahlLoader()

    expect(getArbeitsmodus()).toBe('tischservice')
  })

  it('Direktverkauf-Loader persistiert Direktverkauf', () => {
    ServiceDirektverkaufLoader()

    expect(getArbeitsmodus()).toBe('direktverkauf')
  })
})
