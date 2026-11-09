import { beforeEach, describe, expect, it } from 'vitest'

import { signIn, signOut } from '@/test/auth'

import { getArbeitsmodus } from './lib/arbeitsmodus'
import type { UserRole } from './lib/identity'
import {
  AdminGuard,
  AuthRedirect,
  ServiceDirektverkaufLoader,
  ServiceGuard,
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
    ServiceDirektverkaufLoader()

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
    ServiceDirektverkaufLoader()
    ServiceTischauswahlLoader()

    expect(locationOf(ServiceIndexRedirect())).toBe('tische')
  })
})

function anmelden(role: UserRole | 'none') {
  if (role === 'none') signOut()
  else signIn({ role })
}

function zielVon(result: unknown): string | undefined {
  return result === undefined ? undefined : locationOf(result)
}

describe('Rollen-Guards', () => {
  beforeEach(() => {
    localStorage.clear()
    signOut()
  })

  it.each([
    ['none', '/'],
    ['service', '/'],
    ['serviceleitung', '/'],
    ['admin', undefined],
  ] as const)('AdminGuard für %s leitet auf %s', (role, ziel) => {
    anmelden(role)

    expect(zielVon(AdminGuard())).toBe(ziel)
  })

  it.each([
    ['none', '/'],
    ['service', undefined],
  ] as const)('ServiceGuard für %s leitet auf %s', (role, ziel) => {
    anmelden(role)

    expect(zielVon(ServiceGuard())).toBe(ziel)
  })

  it.each([
    ['none', undefined],
    ['service', '/service'],
    ['serviceleitung', '/service'],
    ['admin', '/admin'],
  ] as const)('AuthRedirect für %s leitet auf %s', (role, ziel) => {
    anmelden(role)

    expect(zielVon(AuthRedirect())).toBe(ziel)
  })
})
