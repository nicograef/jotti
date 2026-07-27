import { z } from 'zod'

import { DateStringSchema, PositionRefSchema } from '../schemas'
import { PositionSchema } from './Bestellung'

export const StornierungSchema = z.object({
  art: z.literal('stornierung'),
  id: z.uuid(),
  userId: z.number().int().min(1),
  // Eingefrorener Name des Akteurs, beschriftet den Historien-Eintrag.
  userName: z.string().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtStornierungCents: z.number().int().min(0),
  // Response-Schema ohne min: die geldneutrale Korrektur hat einen leeren
  // Kommentar (Eingabepflicht min 3 gilt nur im Request, s. u.).
  kommentar: z.string().max(100),
  // barRueckgabe unterscheidet die kassenwirksame Warenrücknahme (true) von der
  // geldneutralen Korrektur (false); vom Backend aus dem Event-Typ abgeleitet.
  barRueckgabe: z.boolean(),
  storniertAm: DateStringSchema,
})
export type Stornierung = z.infer<typeof StornierungSchema>

export const StornierungErteilenSchema = z.object({
  // Client-erzeugter Schlüssel des fachlichen Vorgangs. Ein Wiederholversuch
  // nach Verbindungsabbruch trägt denselben Schlüssel und bucht kein zweites Mal.
  vorgangId: z.uuid(),
  tischId: z.number().int().min(1),
  positionen: PositionRefSchema.array().min(1),
  kommentar: z.string().min(3).max(100),
})
export type StornierungErteilen = z.infer<typeof StornierungErteilenSchema>
