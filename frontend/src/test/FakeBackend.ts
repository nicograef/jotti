import type { z } from 'zod'

import {
  type BackendClient,
  BackendError,
  type DownloadResult,
  ResponseBodyError,
} from '@/lib/Backend'

type Handler = (body: unknown) => unknown

export interface Call {
  endpoint: string
  body: unknown
}

/**
 * In-memory BackendClient that answers per endpoint. A response runs through
 * the caller's schema like a real one, so fixtures must match the wire format.
 */
export class FakeBackend implements BackendClient {
  private readonly handlers = new Map<string, Handler>()
  private readonly downloads = new Map<string, () => DownloadResult>()
  public readonly calls: Call[] = []

  /** A function response receives the request body and may throw or be a vi.fn. */
  public respond(endpoint: string, response: unknown): this {
    this.handlers.set(
      endpoint,
      typeof response === 'function' ? (response as Handler) : () => response,
    )
    return this
  }

  /** A 4xx is not retried by createQueryClient, so the failure shows at once. */
  public fail(
    endpoint: string,
    error: Error = new BackendError(400, 'test_fehler'),
  ): this {
    return this.respond(endpoint, () => {
      throw error
    })
  }

  public respondDownload(endpoint: string, result: DownloadResult): this {
    this.downloads.set(endpoint, () => result)
    return this
  }

  public bodies(endpoint: string): unknown[] {
    return this.calls
      .filter((call) => call.endpoint === endpoint)
      .map((call) => call.body)
  }

  public async post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse> {
    this.calls.push({ endpoint, body })
    const handler = this.handlers.get(endpoint)
    if (!handler) {
      throw new BackendError(404, 'kein_fake', endpoint)
    }

    const response: unknown = await handler(body)
    if (!responseSchema) {
      return {} as TResponse
    }

    const { error, data } = responseSchema.safeParse(response)
    if (error) {
      throw new ResponseBodyError(
        `Fake response of ${endpoint} is invalid: ${error.message}`,
      )
    }
    return data
  }

  public async download(
    endpoint: string,
    body: unknown,
  ): Promise<DownloadResult> {
    this.calls.push({ endpoint, body })
    const download = this.downloads.get(endpoint)
    if (!download) {
      throw new BackendError(404, 'kein_fake', endpoint)
    }
    return Promise.resolve(download())
  }
}
