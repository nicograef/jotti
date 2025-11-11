import { jwtDecode } from 'jwt-decode'
import { z } from 'zod'

const JottiTokenSchema = z.object({
  iss: z.literal('jotti'),
  exp: z.int().min(0),
  iat: z.int().min(0),
  sub: z.string().min(1),
  role: z.enum(['admin', 'service']),
})
type JottiToken = z.infer<typeof JottiTokenSchema>

class Auth {
  private token: JottiToken | null = null
  private tokenBase64: string | null = null

  public get isAuthenticated(): boolean {
    const tokenBase64 = localStorage.getItem('JOTTI_TOKEN')
    if (!tokenBase64) return false

    try {
      this.validateAndSetToken(tokenBase64)
      return true
    } catch (error) {
      console.error('Invalid token:', error)
      return false
    }
  }

  public getToken(): string | null {
    if (!this.tokenBase64) {
      const tokenBase64 = localStorage.getItem('JOTTI_TOKEN')
      if (!tokenBase64) return null
      this.validateAndSetToken(tokenBase64)
    }
    return this.tokenBase64
  }

  public get username(): string | null {
    return this.token?.sub ?? null
  }

  public get isAdmin(): boolean {
    return this.token?.role === 'admin'
  }

  public logout(): void {
    localStorage.removeItem('JOTTI_TOKEN')
    this.token = null
    this.tokenBase64 = null
  }

  public validateAndSetToken(tokenBase64: string): void {
    try {
      const token = jwtDecode<JottiToken>(tokenBase64)

      const { error, data: parsedToken } = JottiTokenSchema.safeParse(token)
      if (error) {
        throw new Error(`Token is invalid: ${error.message}`)
      }

      const currentTime = Math.floor(Date.now() / 1000)
      if (parsedToken.exp < currentTime) {
        throw new Error('Token has expired.')
      }

      this.setToken(parsedToken, tokenBase64)
    } catch (error) {
      this.logout()
      throw new Error('Failed to decode or validate token', { cause: error })
    }
  }

  private setToken(token: JottiToken, tokenBase64: string): void {
    this.token = token
    this.tokenBase64 = tokenBase64
    localStorage.setItem('JOTTI_TOKEN', tokenBase64)
  }
}

export const AuthSingleton = new Auth()
