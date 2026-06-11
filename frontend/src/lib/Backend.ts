import { z } from 'zod'

import { AuthSingleton } from './Auth'

const ErrorResponseSchema = z.object({
  code: z.string(),
  details: z.unknown().optional(),
})

export class BackendError extends Error {
  public readonly status: number
  public readonly code: string
  public readonly details?: unknown

  constructor(status: number, code: string, details?: unknown) {
    let errorMessage = `BackendError: ${code}`

    if (details !== undefined) {
      const detailMessage =
        typeof details === 'string' ? details : JSON.stringify(details)
      errorMessage += ` - ${detailMessage}`
    }

    super(errorMessage)
    this.status = status
    this.code = code
    this.details = details
    Object.setPrototypeOf(this, BackendError.prototype)
  }
}

export class ResponseBodyError extends Error {
  constructor(message: string) {
    super(message)
    Object.setPrototypeOf(this, ResponseBodyError.prototype)
  }
}

// TokenGetter abstracts token retrieval from the Backend class.
// This single-implementation interface is intentional: it enables unit testing of Backend
// without a real authentication dependency by injecting a test double.
interface TokenGetter {
  getToken(): string | null
}

function formatIssuePath(path: readonly PropertyKey[]): string {
  if (path.length === 0) {
    return '<root>'
  }

  let out = ''
  for (const segment of path) {
    if (typeof segment === 'number') {
      out += `[${segment.toString()}]`
      continue
    }

    const key = String(segment)
    out += out === '' ? key : `.${key}`
  }

  return out
}

function formatSchemaIssues(error: z.ZodError): string {
  return error.issues
    .map((issue) => `${formatIssuePath(issue.path)} (${issue.code})`)
    .join(', ')
}

function parseJsonSafely(text: string): unknown {
  try {
    return JSON.parse(text)
  } catch {
    return null
  }
}

export interface BackendClient {
  post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse>
}

export class Backend implements BackendClient {
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
      if (response.status === 401) {
        AuthSingleton.logout()
        window.location.href = '/login'
        throw new BackendError(401, 'unauthorized')
      }

      const responseText = await response.text()
      const parsedError = ErrorResponseSchema.safeParse(
        parseJsonSafely(responseText),
      )

      if (parsedError.success) {
        throw new BackendError(
          response.status,
          parsedError.data.code,
          parsedError.data.details,
        )
      }

      console.error('Failed to parse error response:', parsedError.error)
      console.log('Response text:', responseText)
      throw new BackendError(response.status, 'unknown', responseText)
    }

    if (!responseSchema) {
      // No response schema provided, return empty object
      return {} as TResponse
    }

    const { error, data } = responseSchema.safeParse(await response.json())
    if (error) {
      const issues = formatSchemaIssues(error)
      const message = `Response of ${endpoint} is invalid: ${issues}`
      console.error(message)
      throw new ResponseBodyError(message)
    }

    return data
  }
}

export const BackendSingleton = new Backend('/api', AuthSingleton)
