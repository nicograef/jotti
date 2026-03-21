import { z } from 'zod'

import { PositionRefSchema, PositionSchema } from './Bestellung'

export const AusgabeSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  kommentar: z.string().max(100),
  ausgegebenAm: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Ungültiges Datumsformat',
  }),
})
export type Ausgabe = z.infer<typeof AusgabeSchema>

export const AusgabeBestaetigenSchema = z.object({
  tischId: z.number().int().min(1),
  positionen: PositionRefSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type AusgabeBestaetigen = z.infer<typeof AusgabeBestaetigenSchema>
