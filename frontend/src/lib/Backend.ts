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
  public readonly referenz?: string

  constructor(
    status: number,
    code: string,
    details?: unknown,
    referenz?: string,
  ) {
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
    this.referenz = referenz
    Object.setPrototypeOf(this, BackendError.prototype)
  }
}

export class ResponseBodyError extends Error {
  constructor(message: string) {
    super(message)
    Object.setPrototypeOf(this, ResponseBodyError.prototype)
  }
}

// TokenGetter abstracts token retrieval so Backend can be unit-tested with a test double.
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

// DownloadResult is a binary response (e.g. a file download) together with the
// filename the backend proposed via the Content-Disposition header.
export interface DownloadResult {
  blob: Blob
  filename: string
}

export interface BackendClient {
  post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse>
  download(endpoint: string, body: unknown): Promise<DownloadResult>
}

function parseFilename(contentDisposition: string | null): string {
  if (!contentDisposition) {
    return 'download'
  }
  const match = /filename="?([^"]+)"?/.exec(contentDisposition)
  return match?.[1] ?? 'download'
}

export class Backend implements BackendClient {
  private readonly baseUrl: string
  private readonly tokenGetter: TokenGetter

  constructor(baseUrl: string, tokenGetter: TokenGetter) {
    this.baseUrl = baseUrl
    this.tokenGetter = tokenGetter
  }

  private async request(endpoint: string, body: unknown): Promise<Response> {
    const token = this.tokenGetter.getToken()
    return fetch(`${this.baseUrl}/${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(body),
    })
  }

  private async throwIfNotOk(response: Response): Promise<void> {
    if (response.ok) {
      return
    }

    if (response.status === 401) {
      AuthSingleton.logout()
      window.location.href = '/login'
      throw new BackendError(401, 'unauthorized')
    }

    const referenz = response.headers.get('X-Correlation-ID') ?? undefined
    const responseText = await response.text()
    const parsedError = ErrorResponseSchema.safeParse(
      parseJsonSafely(responseText),
    )

    if (parsedError.success) {
      throw new BackendError(
        response.status,
        parsedError.data.code,
        parsedError.data.details,
        referenz,
      )
    }

    console.error(
      'Failed to parse error response:',
      parsedError.error,
      'Response text:',
      responseText,
    )
    throw new BackendError(response.status, 'unknown', responseText, referenz)
  }

  public async post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse> {
    const response = await this.request(endpoint, body)
    await this.throwIfNotOk(response)

    if (!responseSchema) {
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

  public async download(
    endpoint: string,
    body: unknown,
  ): Promise<DownloadResult> {
    const response = await this.request(endpoint, body)
    await this.throwIfNotOk(response)

    const blob = await response.blob()
    const filename = parseFilename(response.headers.get('Content-Disposition'))
    return { blob, filename }
  }
}

export const BackendSingleton = new Backend('/api', AuthSingleton)
