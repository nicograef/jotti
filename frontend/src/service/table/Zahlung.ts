import { z } from 'zod'

import { PositionRefSchema, PositionSchema } from './Bestellung'

export const ZahlungSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtZahlungCents: z.number().int().min(0),
  kommentar: z.string().max(100),
  registriertAm: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Zahlung = z.infer<typeof ZahlungSchema>

export const ZahlungRegistrierenSchema = z.object({
  tischId: z.number().int().min(1),
  positionen: PositionRefSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type ZahlungRegistrieren = z.infer<typeof ZahlungRegistrierenSchema>
