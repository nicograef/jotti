import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

const HealthResponseSchema = z.object({
  version: z.string(),
})

export class HealthBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getVersion(): Promise<string> {
    const { version } = await this.backend.post(
      'health',
      {},
      HealthResponseSchema,
    )
    return version
  }
}
