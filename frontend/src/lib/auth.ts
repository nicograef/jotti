import { jwtDecode } from "jwt-decode"
import { z } from "zod"
import { BackendSingleton } from "./backend"

const JottiTokenSchema = z.object({
  iss: z.literal("jotti"),
  exp: z.int().min(0),
  iat: z.int().min(0),
  sub: z.string().min(1),
  role: z.enum(["admin", "service"]),
})
type JottiToken = z.infer<typeof JottiTokenSchema>

interface Backend {
  login(username: string, password: string): Promise<string>
}

class Auth {
  private backend: Backend
  private token: JottiToken | null = null

  constructor(backend: Backend) {
    this.backend = backend
  }

  public get isAuthenticated(): boolean {
    const token = localStorage.getItem("JOTTI_TOKEN")
    if (!token) return false

    try {
      this.validateAndSetToken(token)
      return true
    } catch (error) {
      console.error("Invalid token:", error)
      return false
    }
  }

  public get username(): string | null {
    return this.token?.sub ?? null
  }

  public get isAdmin(): boolean {
    return this.token?.role === "admin"
  }

  public async login(username: string, password: string): Promise<void> {
    const token = await this.backend.login(username, password)
    this.validateAndSetToken(token)
    localStorage.setItem("JOTTI_TOKEN", token)
  }

  private validateAndSetToken(tokenBase64: string): void {
    try {
      const token = jwtDecode<JottiToken>(tokenBase64)

      const { error, data: parsedToken } = JottiTokenSchema.safeParse(token)
      if (error) {
        throw new Error(`Token is invalid: ${error.message}`)
      }

      const currentTime = Math.floor(Date.now() / 1000)
      if (parsedToken.exp < currentTime) {
        throw new Error("Token has expired.")
      }

      this.token = parsedToken
    } catch (error) {
      localStorage.removeItem("JOTTI_TOKEN")
      this.token = null
      throw new Error("Failed to decode or validate token", { cause: error })
    }
  }
}

export const AuthSingleton = new Auth(BackendSingleton)
