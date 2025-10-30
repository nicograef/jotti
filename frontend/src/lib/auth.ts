import { jwtDecode } from "jwt-decode"
import { z } from "zod"
import { BackendSingleton } from "./backend"
import { redirect } from "react-router"

const JottiTokenSchema = z.object({
  iss: z.literal("jotti"),
  exp: z.int().min(1750000000000), // some date newer than 2025
  sub: z.string().min(4),
  roles: z.array(z.string()),
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
    const token = localStorage.getItem("JOTTI_TOKEN")
    if (token) this.validateAndSetToken(token)
  }

  public get isAuthenticated(): boolean {
    if (!this.token) return false
    const currentTime = Math.floor(Date.now() / 1000)
    return this.token.exp ? this.token.exp > currentTime : true
  }

  public get username(): string | null {
    return this.token?.sub ?? null
  }

  public get isAdmin(): boolean {
    return this.token?.roles.includes("admin") ?? false
  }

  public async login(username: string, password: string): Promise<void> {
    const token = await this.backend.login(username, password)
    this.validateAndSetToken(token)

    if (this.isAdmin) {
      redirect("/admin")
    } else {
      redirect("/")
    }
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
      localStorage.setItem("JOTTI_TOKEN", JSON.stringify(tokenBase64))
    } catch (error) {
      localStorage.removeItem("JOTTI_TOKEN")
      this.token = null
      throw new Error("Failed to decode or validate token", { cause: error })
    }
  }
}

export const AuthSingleton = new Auth(BackendSingleton)
