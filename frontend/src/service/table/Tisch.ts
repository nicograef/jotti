import { z } from 'zod'

import { TischIdSchema } from '../schemas'
import { BestellungSchema, PositionSchema } from './Bestellung'
import { StornierungSchema } from './Stornierung'
import { UmbuchungSchema } from './Umbuchung'
import { ZahlungSchema } from './Zahlung'

export interface Tisch {
  id: number
  name: string
  saldoCents: number
}

export const TischSessionSchema = z.object({
  tischId: TischIdSchema,
  tischName: z.string(),
  saldoCents: z.number().int(),
  unbezahltePositionen: z.array(PositionSchema),
  // True, wenn die anfragende Servicekraft an diesem Tisch keine eigenen
  // unbezahlten Positionen mehr hat.
  fuerMichErledigt: z.boolean(),
})
export type TischSession = z.infer<typeof TischSessionSchema>

export const TischHistorieEintragSchema = z.discriminatedUnion('art', [
  BestellungSchema,
  ZahlungSchema,
  StornierungSchema,
  UmbuchungSchema,
])
export type TischHistorieEintrag = z.infer<typeof TischHistorieEintragSchema>

export const AktiverTischMitFavoritSchema = z.object({
  id: TischIdSchema,
  name: z.string(),
  saldoCents: z.number().int(),
  istFavorit: z.boolean(),
})
export type AktiverTischMitFavorit = z.infer<
  typeof AktiverTischMitFavoritSchema
>

// anzahlRuecknahmen/ruecknahmenCents sind die dieser Servicekraft zugeordneten
// Warenrücknahmen (zurückgenommen von einer Zahlung, die sie kassiert hat — egal, wer
// storniert hat). pfandCents ist das am Tisch ausgezahlte Pfand. abzugebenCents ist
// zahlungenCents abzüglich Rücknahmen und Pfand und darf negativ werden.
export const EigeneUebersichtSchema = z.object({
  anzahlBestellungen: z.number().int(),
  bestellungenCents: z.number().int(),
  anzahlZahlungen: z.number().int(),
  zahlungenCents: z.number().int(),
  anzahlRuecknahmen: z.number().int(),
  ruecknahmenCents: z.number().int(),
  pfandCents: z.number().int(),
  abzugebenCents: z.number().int(),
})
export type EigeneUebersicht = z.infer<typeof EigeneUebersichtSchema>
