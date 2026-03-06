import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { AuthSingleton } from './Auth'

// Helper to create a valid JWT-like token (header.payload.signature)
function createMockToken(payload: Record<string, unknown>): string {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  const body = btoa(JSON.stringify(payload))
  const signature = btoa('fake-signature')
  return `${header}.${body}.${signature}`
}

function validTokenPayload(overrides: Record<string, unknown> = {}) {
  return {
    iss: 'jotti',
    exp: Math.floor(Date.now() / 1000) + 3600, // 1h from now
    iat: Math.floor(Date.now() / 1000),
    sub: 1,
    role: 'admin',
    ...overrides,
  }
}

describe('Auth', () => {
  beforeEach(() => {
    localStorage.clear()
    AuthSingleton.logout()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('validateAndSetToken', () => {
    it('accepts a valid token', () => {
      const token = createMockToken(validTokenPayload())

      expect(() => {
        AuthSingleton.validateAndSetToken(token)
      }).not.toThrow()
      expect(AuthSingleton.getToken()).toBe(token)
      expect(localStorage.getItem('JOTTI_TOKEN')).toBe(token)
    })

    it('rejects an expired token', () => {
      const token = createMockToken(
        validTokenPayload({ exp: Math.floor(Date.now() / 1000) - 100 }),
      )

      expect(() => {
        AuthSingleton.validateAndSetToken(token)
      }).toThrow()
      expect(AuthSingleton.getToken()).toBeNull()
    })

    it('rejects a token with wrong issuer', () => {
      const token = createMockToken(validTokenPayload({ iss: 'other' }))

      expect(() => {
        AuthSingleton.validateAndSetToken(token)
      }).toThrow()
    })

    it('rejects a token with invalid role', () => {
      const token = createMockToken(validTokenPayload({ role: 'superadmin' }))

      expect(() => {
        AuthSingleton.validateAndSetToken(token)
      }).toThrow()
    })
  })

  describe('isAuthenticated', () => {
    it('returns false when no token exists', () => {
      expect(AuthSingleton.isAuthenticated).toBe(false)
    })

    it('returns true with valid token in localStorage', () => {
      const token = createMockToken(validTokenPayload())
      localStorage.setItem('JOTTI_TOKEN', token)

      expect(AuthSingleton.isAuthenticated).toBe(true)
    })

    it('returns false and clears expired token', () => {
      const token = createMockToken(
        validTokenPayload({ exp: Math.floor(Date.now() / 1000) - 100 }),
      )
      localStorage.setItem('JOTTI_TOKEN', token)

      expect(AuthSingleton.isAuthenticated).toBe(false)
      expect(localStorage.getItem('JOTTI_TOKEN')).toBeNull()
    })
  })

  describe('role checks', () => {
    it('detects admin role', () => {
      const token = createMockToken(validTokenPayload({ role: 'admin' }))
      AuthSingleton.validateAndSetToken(token)

      expect(AuthSingleton.isAdmin).toBe(true)
      expect(AuthSingleton.isSeniorService).toBe(false)
      expect(AuthSingleton.isService).toBe(false)
      expect(AuthSingleton.canCancel).toBe(true)
    })

    it('detects senior_service role', () => {
      const token = createMockToken(
        validTokenPayload({ role: 'senior_service' }),
      )
      AuthSingleton.validateAndSetToken(token)

      expect(AuthSingleton.isAdmin).toBe(false)
      expect(AuthSingleton.isSeniorService).toBe(true)
      expect(AuthSingleton.canCancel).toBe(true)
    })

    it('detects service role', () => {
      const token = createMockToken(validTokenPayload({ role: 'service' }))
      AuthSingleton.validateAndSetToken(token)

      expect(AuthSingleton.isService).toBe(true)
      expect(AuthSingleton.canCancel).toBe(false)
    })
  })

  describe('logout', () => {
    it('clears token state and localStorage', () => {
      const token = createMockToken(validTokenPayload())
      AuthSingleton.validateAndSetToken(token)

      AuthSingleton.logout()

      expect(AuthSingleton.getToken()).toBeNull()
      expect(AuthSingleton.userId).toBeNull()
      expect(localStorage.getItem('JOTTI_TOKEN')).toBeNull()
    })
  })

  describe('userId', () => {
    it('returns userId from valid token', () => {
      const token = createMockToken(validTokenPayload({ sub: 42 }))
      AuthSingleton.validateAndSetToken(token)

      expect(AuthSingleton.userId).toBe(42)
    })

    it('returns null when no token is set', () => {
      expect(AuthSingleton.userId).toBeNull()
    })
  })

  describe('getToken', () => {
    it('loads token from localStorage if not cached', () => {
      const token = createMockToken(validTokenPayload())

      // Clear internal state, then put token in localStorage
      AuthSingleton.logout()
      localStorage.setItem('JOTTI_TOKEN', token)

      // getToken should read from localStorage
      const result = AuthSingleton.getToken()
      expect(result).toBe(token)
    })

    it('returns null when localStorage is empty', () => {
      expect(AuthSingleton.getToken()).toBeNull()
    })
  })
})
