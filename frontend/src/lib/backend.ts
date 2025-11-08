import { z } from "zod"
import {
  CreateUserRequestSchema,
  CreateUserResponseSchema,
  GetUsersResponseSchema,
  LoginRequestSchema,
  LoginResponseSchema,
  SetPasswordRequestSchema,
  SetPasswordResponseSchema,
  UpdateUserRequestSchema,
  UpdateUserResponseSchema,
  type User,
  type UserRole,
} from "./user"
import { AuthSingleton } from "./auth"

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

interface TokenGetter {
  getToken(): string | null
}

class Backend {
  private readonly baseUrl: string
  private readonly tokenGetter: TokenGetter

  constructor(baseUrl: string, tokenGetter: TokenGetter) {
    this.baseUrl = baseUrl
    this.tokenGetter = tokenGetter
  }

  /** Sends a login request with the given username and password and returns the JWT token from the backend. */
  public async login(username: string, password: string): Promise<string> {
    const body = LoginRequestSchema.parse({ username, password })
    const { token } = await this.post("login", body, LoginResponseSchema)
    return token
  }

  /** Sends a login request with the given username and password and returns the JWT token from the backend. */
  public async setPassword(
    username: string,
    password: string,
    onetimePassword: string,
  ): Promise<string> {
    const body = SetPasswordRequestSchema.parse({
      username,
      password,
      onetimePassword,
    })
    const { token } = await this.post(
      "set-password",
      body,
      SetPasswordResponseSchema,
    )
    return token
  }

  public async createUser(
    name: string,
    username: string,
    role: UserRole,
  ): Promise<{ user: User; onetimePassword: string }> {
    const body = CreateUserRequestSchema.parse({ name, username, role })
    const { user, onetimePassword } = await this.post(
      "admin/create-user",
      body,
      CreateUserResponseSchema,
    )
    return { user, onetimePassword }
  }

  public async updateUser(
    updatedUser: z.infer<typeof UpdateUserRequestSchema>,
  ): Promise<User> {
    const body = UpdateUserRequestSchema.parse(updatedUser)
    const { user } = await this.post(
      "admin/update-user",
      body,
      UpdateUserResponseSchema,
    )
    return user
  }

  public async getUsers(): Promise<User[]> {
    const { users } = await this.post(
      "admin/get-users",
      {},
      GetUsersResponseSchema,
    )
    return users
  }

  private async post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema: z.ZodType<TResponse>,
  ): Promise<TResponse> {
    const token = this.tokenGetter.getToken()
    const response = await fetch(`${this.baseUrl}/${endpoint}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
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
  AuthSingleton,
)
