import { z } from 'zod'

import { DateStringSchema } from '@/lib/utils'

import {
  GebuchtePositionSchema,
  KommentarSchema,
  PositionInputSchema,
  TischIdSchema,
} from '../schemas'

export const PositionSchema = GebuchtePositionSchema.extend({
  // Besteller (aufnehmende Servicekraft). In der Tisch-Session gesetzt; in der
  // Historie (stornierbare/umbuchbare Positionen) leer (0/'').
  bestellerUserId: z.number().int(),
  bestellerName: z.string(),
})
export type Position = z.infer<typeof PositionSchema>

export const BestellungAufnehmenSchema = z.object({
  bestellungId: z.uuid(),
  tischId: TischIdSchema,
  positionen: PositionInputSchema.array().min(1),
  kommentar: KommentarSchema,
})
export type BestellungAufnehmen = z.infer<typeof BestellungAufnehmenSchema>

export const BestellungSchema = z.object({
  art: z.literal('bestellung'),
  id: z.uuid(),
  userId: z.number().int().min(1),
  // Eingefrorener Name der bestellenden Servicekraft.
  userName: z.string().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtPreisCents: z.number().int().min(0),
  kommentar: KommentarSchema,
  aufgenommenAm: DateStringSchema,
  // Backend-computed (single source of truth): positions of this order that are
  // still stornierbar (ordered − cancelled − moved away) resp. umbuchbar
  // (− paid as well).
  stornierbarePositionen: PositionSchema.array(),
  umbuchbarePositionen: PositionSchema.array(),
})
export type Bestellung = z.infer<typeof BestellungSchema>
