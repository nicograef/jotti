import { z } from 'zod'

export const PositionSchema = z.object({
  id: z.number().int().min(1),
  name: z.string().min(1).max(100),
  preisCents: z.number().int().min(0),
  menge: z.number().int().min(1),
})
export type Position = z.infer<typeof PositionSchema>

export const BestellungAufgebenSchema = z.object({
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type BestellungAufgeben = z.infer<typeof BestellungAufgebenSchema>

export const BestellungSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtPreisCents: z.number().int().min(0),
  kommentar: z.string().max(100),
  aufgegebenAm: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Bestellung = z.infer<typeof BestellungSchema>
