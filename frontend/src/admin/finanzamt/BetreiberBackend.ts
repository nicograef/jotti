import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import { DateStringSchema } from '@/lib/utils'

// Die Adressfelder des Betreibers (Eingabe beim Speichern). Sie erscheinen auf
// jedem Kassenbeleg (§ 6 KassenSichV).
export const BetreiberEingabeSchema = z.object({
  vereinsname: z.string(),
  strasse: z.string(),
  plz: z.string(),
  ort: z.string(),
  steuernummer: z.string().nullable(),
  ustId: z.string().nullable(),
})
export type BetreiberEingabe = z.infer<typeof BetreiberEingabeSchema>

// Der Betreiber wie ihn die Query liefert: Adressfelder plus der Status der
// ELSTER-Kassenmeldung (Datum als YYYY-MM-DD oder null, solange noch nicht
// gemeldet, § 146a Abs. 4 AO).
export const BetreiberSchema = BetreiberEingabeSchema.extend({
  elsterGemeldetAm: DateStringSchema.nullable(),
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

  public async saveBetreiber(betreiber: BetreiberEingabe): Promise<void> {
    const body = BetreiberEingabeSchema.parse(betreiber)
    await this.backend.post('admin/update-betreiber', body)
  }

  public async setElsterMeldung(): Promise<void> {
    await this.backend.post('admin/elster-meldung-setzen', {})
  }

  public async nimmElsterMeldungZurueck(): Promise<void> {
    await this.backend.post('admin/elster-meldung-zuruecknehmen', {})
  }
}
