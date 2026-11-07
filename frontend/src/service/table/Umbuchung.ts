import { z } from 'zod'

import { DateStringSchema } from '@/lib/utils'

import { KommentarSchema, PositionRefSchema, TischIdSchema } from '../schemas'
import { PositionSchema } from './Bestellung'

export const BestellungUmbuchenSchema = z.object({
  quellTischId: TischIdSchema,
  zielTischId: TischIdSchema,
  positionen: PositionRefSchema.array().min(1),
  benutzerKommentar: KommentarSchema,
})
export type BestellungUmbuchen = z.infer<typeof BestellungUmbuchenSchema>

// History entry of a cash-neutral Umbuchung, shown on the source (Abgang) and target (Zugang) Tisch;
// tischId vs. quellTischId/zielTischId gives the direction. Only the Zugang carries stornier-/umbuchbare positions.
export const UmbuchungSchema = z.object({
  art: z.literal('umbuchung'),
  id: z.uuid(),
  userId: z.number().int().min(1),
  // Eingefrorener Name des Akteurs, beschriftet den Historien-Eintrag.
  userName: z.string().min(1),
  tischId: z.number().int().min(1),
  quellTischId: z.number().int().min(1),
  zielTischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtCents: z.number().int().min(0),
  // kommentar ist der Richtungs-Autotext, benutzerKommentar der optionale freie Text.
  kommentar: KommentarSchema,
  benutzerKommentar: KommentarSchema,
  umgebuchtAm: DateStringSchema,
  // Backend-computed (single source of truth); nur für den Zugang befüllt.
  stornierbarePositionen: PositionSchema.array(),
  umbuchbarePositionen: PositionSchema.array(),
})
export type Umbuchung = z.infer<typeof UmbuchungSchema>
