import { z } from 'zod'

export const PositionSchema = z.object({
  positionId: z.uuid(),
  varianteId: z.number().int().min(1),
  produktName: z.string().min(1).max(100),
  varianteName: z.string().min(1).max(100),
  // The backend sends kategorie as a free-form string; this enum validates it
  // strictly. Adding a new category requires updating this enum in lockstep.
  kategorie: z.enum(['essen', 'getraenk', 'sonstiges']),
  steuersatz: z.enum(['regel', 'ermaessigt', 'befreit', 'kombi']),
  einzelpreis: z.number().int().min(0),
  menge: z.number().int().min(1),
})
export type Position = z.infer<typeof PositionSchema>

export const PositionRefSchema = z.object({
  positionId: z.uuid(),
  menge: z.number().int().min(1),
})
export type PositionRef = z.infer<typeof PositionRefSchema>

export const BestellPositionInputSchema = z.object({
  produktId: z.number().int().min(1),
  varianteId: z.number().int().min(1),
  menge: z.number().int().min(1),
})
export type BestellPositionInput = z.infer<typeof BestellPositionInputSchema>

export const BestellungAufnehmenSchema = z.object({
  tischId: z.number().int().min(1),
  positionen: BestellPositionInputSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type BestellungAufnehmen = z.infer<typeof BestellungAufnehmenSchema>

export const BestellungSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtPreisCents: z.number().int().min(0),
  kommentar: z.string().max(100),
  aufgenommenAm: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Ungültiges Datumsformat',
  }),
})
export type Bestellung = z.infer<typeof BestellungSchema>
