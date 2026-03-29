import { z } from 'zod'

import { PositionSchema } from './Bestellung'

export const TischIdSchema = z.number().int().min(1)
const TischNameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(100, { message: 'Der Name ist zu lang.' })

export const TischSchema = z.object({
  id: TischIdSchema,
  name: TischNameSchema,
  saldoCents: z.number().int(),
})
export type Tisch = z.infer<typeof TischSchema>

export const TischSessionSchema = z.object({
  tischId: TischIdSchema,
  tischName: z.string(),
  saldoCents: z.number().int(),
  unbezahltePositionen: z.array(PositionSchema),
  ausstehendePositionen: z.array(PositionSchema),
  gesamtZahlungenCents: z.number().int(),
})
export type TischSession = z.infer<typeof TischSessionSchema>

export const AktiverTischMitFavoritSchema = z.object({
  id: TischIdSchema,
  name: z.string(),
  saldoCents: z.number().int(),
  istFavorit: z.boolean(),
})
export type AktiverTischMitFavorit = z.infer<
  typeof AktiverTischMitFavoritSchema
>

export const EigeneUebersichtSchema = z.object({
  anzahlBestellungen: z.number().int(),
  bestellungenCents: z.number().int(),
  anzahlZahlungen: z.number().int(),
  zahlungenCents: z.number().int(),
})
export type EigeneUebersicht = z.infer<typeof EigeneUebersichtSchema>
