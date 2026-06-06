import { z } from 'zod'

import type { BackendClient } from './Backend'

export const BetreiberSchema = z.object({
  vereinsname: z.string(),
  strasse: z.string(),
  plz: z.string(),
  ort: z.string(),
  steuernummer: z.string().nullable(),
  ustId: z.string().nullable(),
})
export type Betreiber = z.infer<typeof BetreiberSchema>

const KassenidentitaetSchema = z.object({
  seriennummer: z.uuid(),
  angelegtAm: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Ungültiges Datumsformat',
  }),
})
export type Kassenidentitaet = z.infer<typeof KassenidentitaetSchema>

export class EinstellungenBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getKassenidentitaet(): Promise<Kassenidentitaet> {
    return this.backend.post(
      'admin/get-kassenidentitaet',
      {},
      KassenidentitaetSchema,
    )
  }

  public async getBetreiber(): Promise<Betreiber> {
    return this.backend.post('admin/get-betreiber', {}, BetreiberSchema)
  }

  public async saveBetreiber(b: Betreiber): Promise<void> {
    const body = BetreiberSchema.parse(b)
    await this.backend.post('admin/update-betreiber', body)
  }
}
