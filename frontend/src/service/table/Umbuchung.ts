import { z } from 'zod'

import { DateStringSchema, PositionRefSchema } from '../schemas'
import { PositionSchema } from './Bestellung'

export const BestellungUmbuchenSchema = z.object({
  quellTischId: z.number().int().min(1),
  zielTischId: z.number().int().min(1),
  positionen: PositionRefSchema.array().min(1),
  benutzerKommentar: z.string().max(100),
})
export type BestellungUmbuchen = z.infer<typeof BestellungUmbuchenSchema>

// Historien-Eintrag einer geldneutralen Umbuchung. Er erscheint sowohl auf dem
// Quelltisch (Abgang) als auch auf dem Zieltisch (Zugang); die Richtung folgt aus
// dem Verhältnis von tischId zu quellTischId/zielTischId. Nur der Zugang trägt
// stornier-/umbuchbare Positionen (er bringt die Positionen auf den Tisch).
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
  kommentar: z.string().max(100),
  benutzerKommentar: z.string().max(100),
  umgebuchtAm: DateStringSchema,
  // Backend-computed (single source of truth); nur für den Zugang befüllt.
  stornierbarePositionen: PositionSchema.array(),
  umbuchbarePositionen: PositionSchema.array(),
})
export type Umbuchung = z.infer<typeof UmbuchungSchema>
