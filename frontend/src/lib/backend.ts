import { z } from "zod"

const LoginResponseSchema = z.object({
  token: z.string().min(10), // validation is done in Auth Service
})

const ErrorResponseSchema = z.object({
  code: z.string(),
  message: z.string(),
})

export class BackendError extends Error {
  public readonly status: number
  public readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.status = status
    this.code = code
    Object.setPrototypeOf(this, BackendError.prototype)
  }
}

export class ResponseBodyError extends Error {
  constructor(message: string) {
    super(message)
    Object.setPrototypeOf(this, ResponseBodyError.prototype)
  }
}

class Backend {
  private baseUrl: string

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }

  /** Sends a login request with the given username and password and returns the JWT token from the backend. */
  public async login(username: string, password: string): Promise<string> {
    const { token } = await this.post(
      "login",
      { username, password },
      LoginResponseSchema,
    )
    return token
  }

  /** Sends a login request with the given username and password and returns the JWT token from the backend. */
  public async setPassword(
    username: string,
    password: string,
    onetimePassword: string,
  ): Promise<string> {
    const { token } = await this.post(
      "set-password",
      { username, password, onetimePassword },
      LoginResponseSchema,
    )
    return token
  }

  private async post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema: z.ZodType<TResponse>,
  ): Promise<TResponse> {
    const response = await fetch(`${this.baseUrl}/${endpoint}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    })

    if (!response.ok) {
      try {
        const responseBody = ErrorResponseSchema.parse(await response.json())
        throw new BackendError(
          response.status,
          responseBody.code,
          responseBody.message,
        )
      } catch (error) {
        if (error instanceof BackendError) throw error

        console.error("Failed to parse error response:", error)
        const responseText = await response.text()
        console.log("Response text:", responseText)
        throw new BackendError(response.status, "unknown", responseText)
      }
    }

    const { error, data } = responseSchema.safeParse(await response.json())
    if (error) {
      console.error(error.message)
      throw new ResponseBodyError(`Response of ${endpoint} is invalid`)
    }

    return data
  }
}

export const BackendSingleton = new Backend(
  "https://automatic-space-umbrella-v655jg4vp5jfp69-3000.app.github.dev",
)
