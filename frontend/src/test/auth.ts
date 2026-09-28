import { AuthSingleton } from '@/lib/Auth'

type Role = 'admin' | 'serviceleitung' | 'service'

function base64Url(value: object): string {
  return btoa(JSON.stringify(value))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
}

/** Signs in with an unsigned token; the client decodes JWTs without verifying them. */
export function signIn({
  userId = 1,
  role = 'service',
}: { userId?: number; role?: Role } = {}): void {
  const now = Math.floor(Date.now() / 1000)
  const payload = { iss: 'jotti', iat: now, exp: now + 3600, sub: userId, role }
  const token = `${base64Url({ alg: 'none', typ: 'JWT' })}.${base64Url(payload)}.`
  AuthSingleton.validateAndSetToken(token)
}

export function signOut(): void {
  AuthSingleton.logout()
}
