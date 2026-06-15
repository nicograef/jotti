import { z } from 'zod'

import { KategorieSchema } from '../product/Produkt'
import { DateStringSchema, SteuersatzSchema } from '../schemas'

export const PositionSchema = z.object({
  positionId: z.uuid(),
  varianteId: z.number().int().min(1),
  produktName: z.string().min(1).max(100),
  varianteName: z.string().min(1).max(100),
  kategorie: KategorieSchema,
  steuersatz: SteuersatzSchema,
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
  art: z.literal('bestellung'),
  id: z.uuid(),
  userId: z.number().int().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtPreisCents: z.number().int().min(0),
  kommentar: z.string().max(100),
  aufgenommenAm: DateStringSchema,
  // Backend-computed (single source of truth): positions of this order that are
  // still stornierbar (ordered − cancelled) resp. umbuchbar (− paid as well).
  stornierbarePositionen: PositionSchema.array(),
  umbuchbarePositionen: PositionSchema.array(),
})
export type Bestellung = z.infer<typeof BestellungSchema>
