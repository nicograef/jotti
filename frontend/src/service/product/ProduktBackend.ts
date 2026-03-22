import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import { type Produkt, ProduktSchema } from './Produkt'

export class ProduktBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getAktiveProdukte(): Promise<Produkt[]> {
    const { produkte } = await this.backend.post(
      'service/get-aktive-produkte',
      {},
      z.object({ produkte: z.array(ProduktSchema) }),
    )
    return produkte
  }
}
