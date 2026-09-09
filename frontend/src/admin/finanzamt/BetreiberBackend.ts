import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import { DateStringSchema } from '@/lib/utils'

// Die Adressfelder des Betreibers (Eingabe beim Speichern). Sie erscheinen auf
// jedem Kassenbeleg (§ 6 KassenSichV). Grenzen und Trim spiegeln das zog-Schema
// in domain/betreiber (Regel 5): die amtlichen Maximallängen der
// DSFinV-K-Stammdaten.
export const BetreiberEingabeSchema = z.object({
  vereinsname: z
    .string()
    .trim()
    .min(1, { message: 'Vereinsname ist erforderlich.' })
    .max(60, { message: 'Der Vereinsname ist zu lang.' }),
  strasse: z
    .string()
    .trim()
    .min(1, { message: 'Straße ist erforderlich.' })
    .max(60, { message: 'Die Straße ist zu lang.' }),
  plz: z
    .string()
    .trim()
    .min(1, { message: 'PLZ ist erforderlich.' })
    .max(10, { message: 'Die PLZ ist zu lang.' }),
  ort: z
    .string()
    .trim()
    .min(1, { message: 'Ort ist erforderlich.' })
    .max(62, { message: 'Der Ort ist zu lang.' }),
  steuernummer: z
    .string()
    .trim()
    .max(20, { message: 'Die Steuernummer ist zu lang.' })
    .nullable(),
  ustId: z
    .string()
    .trim()
    .max(15, { message: 'Die USt-ID ist zu lang.' })
    .nullable(),
})
export type BetreiberEingabe = z.infer<typeof BetreiberEingabeSchema>

// Der Betreiber wie ihn die Query liefert: Adressfelder plus der Status der
// ELSTER-Kassenmeldung (Datum als YYYY-MM-DD oder null, solange noch nicht
// gemeldet, § 146a Abs. 4 AO). Die Felder tragen hier keine Grenzen: Vor der
// Einrichtung liefert die Query leere Felder, und ein Bestandswert kann länger
// sein als die amtliche Maximallänge — der DSFinV-K-Export kürzt ihn.
export const BetreiberSchema = z.object({
  vereinsname: z.string(),
  strasse: z.string(),
  plz: z.string(),
  ort: z.string(),
  steuernummer: z.string().nullable(),
  ustId: z.string().nullable(),
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
