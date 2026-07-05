import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import { DateStringSchema } from '@/lib/utils'

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
  angelegtAm: DateStringSchema,
})
export type Kassenidentitaet = z.infer<typeof KassenidentitaetSchema>

export class BetreiberBackend {
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

  public async saveBetreiber(betreiber: Betreiber): Promise<void> {
    const body = BetreiberSchema.parse(betreiber)
    await this.backend.post('admin/update-betreiber', body)
  }
}
