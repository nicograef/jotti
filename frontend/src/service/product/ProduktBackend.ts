import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import { ProduktSchema } from '@/lib/produktSchemas'

const AktiveProdukteSchema = z.object({
  produkte: z.array(ProduktSchema),
  pfandProdukte: z.array(ProduktSchema),
})
export type AktiveProdukte = z.infer<typeof AktiveProdukteSchema>

export class ProduktBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  // pfandProdukte ist die Teilmenge mit Pfand-Kennzeichen, gefiltert im Backend.
  public async getAktiveProdukte(): Promise<AktiveProdukte> {
    return this.backend.post(
      'service/get-aktive-produkte',
      {},
      AktiveProdukteSchema,
    )
  }
}
