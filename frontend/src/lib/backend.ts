import { z } from "zod"

const LoginResponseSchema = z.union([
  z.object({
    ok: z.literal(true),
    token: z.string().min(10), // validation is done in Auth Service
  }),
  z.object({
    ok: z.literal(false),
    error: z.string().min(1).max(100),
  }),
])

class Backend {
  private baseUrl: string

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }

  /** Sends a login request with the given username and password and returns the JWT token from the backend. */
  public async login(username: string, password: string): Promise<string> {
    const response = await fetch(`${this.baseUrl}/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Credentials": "true",
        "Access-Control-Allow-Headers": "*",
        "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
      },
      body: JSON.stringify({ username, password }),
    })

    if (!response.ok) {
      throw new Error(`Network response was not ok: ${response.statusText}`)
    }

    const { error, data } = LoginResponseSchema.safeParse(await response.json())
    if (error) {
      console.error(error)
      throw new Error("Login failed")
    } else if (!data.ok) {
      throw new Error(`Login failed: ${data.error}`)
    } else {
      return data.token
    }
  }
}

export const BackendSingleton = new Backend(
  "https://automatic-space-umbrella-v655jg4vp5jfp69-3000.app.github.dev",
)
