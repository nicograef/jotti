import { AuthSingleton } from '@/lib/Auth'

type Role = 'admin' | 'serviceleitung' | 'service'

function base64Url(value: object): string {
  return btoa(JSON.stringify(value))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
}

interface TokenClaims {
  userId?: number
  role?: Role
}

/** An unsigned token; the client decodes JWTs without verifying them. */
export function testToken({
  userId = 1,
  role = 'service',
}: TokenClaims = {}): string {
  const now = Math.floor(Date.now() / 1000)
  const payload = { iss: 'jotti', iat: now, exp: now + 3600, sub: userId, role }
  return `${base64Url({ alg: 'none', typ: 'JWT' })}.${base64Url(payload)}.`
}

export function signIn(claims: TokenClaims = {}): void {
  AuthSingleton.validateAndSetToken(testToken(claims))
}

export function signOut(): void {
  AuthSingleton.logout()
}
