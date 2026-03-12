import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import { type Produkt, ProduktSchema } from './Product'

export class ProductBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getActiveProducts(): Promise<Produkt[]> {
    const { produkte } = await this.backend.post(
      'service/get-aktive-produkte',
      {},
      z.object({ produkte: z.array(ProduktSchema) }),
    )
    return produkte
  }
}
