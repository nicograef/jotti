import { z } from 'zod'

import { AuthSingleton } from './auth'

const ErrorResponseSchema = z.object({
  code: z.string(),
  details: z.string().optional(),
})

export class BackendError extends Error {
  public readonly status: number
  public readonly code: string

  constructor(status: number, code: string, details?: string) {
    super(
      details ? `BackendError: ${code} - ${details}` : `BackendError: ${code}`,
    )
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

  public async post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse> {
    const token = this.tokenGetter.getToken()
    const response = await fetch(`${this.baseUrl}/${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(body),
    })

    if (!response.ok) {
      try {
        const { code, details } = ErrorResponseSchema.parse(
          await response.json(),
        )
        throw new BackendError(response.status, code, details)
      } catch (error) {
        if (error instanceof BackendError) throw error

        console.error('Failed to parse error response:', error)
        const responseText = await response.text()
        console.log('Response text:', responseText)
        throw new BackendError(response.status, 'unknown', responseText)
      }
    }

    if (!responseSchema) {
      // No response schema provided, return empty object
      return {} as TResponse
    }

    const { error, data } = responseSchema.safeParse(await response.json())
    if (error) {
      console.error(error.message)
      throw new ResponseBodyError(`Response of ${endpoint} is invalid`)
    }

    return data
  }
}

export const BackendSingleton = new Backend('/api', AuthSingleton)
