import { z } from 'zod'

import { DateStringSchema } from '@/lib/utils'

import {
  KommentarSchema,
  PflichtKommentarSchema,
  PositionRefSchema,
  TischIdSchema,
} from '../schemas'
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
  // Response-Schema ohne min: das Korrektur-Event (bestellung-korrigiert:v1)
  // lässt einen leeren Kommentar zu (Eingabepflicht min 3 gilt nur im Request,
  // s. u.).
  kommentar: KommentarSchema,
  // barRueckgabe unterscheidet die kassenwirksame Warenrücknahme (true) von der
  // geldneutralen Korrektur (false); vom Backend aus dem Event-Typ abgeleitet.
  barRueckgabe: z.boolean(),
  storniertAm: DateStringSchema,
})
export type Stornierung = z.infer<typeof StornierungSchema>

// stornoId ist der Idempotenz-Schlüssel des Versuchs; das Backend leitet daraus
// jede Event-ID der Stornierung ab.
export const StornierungErteilenSchema = z.object({
  stornoId: z.uuid(),
  tischId: TischIdSchema,
  positionen: PositionRefSchema.array().min(1),
  kommentar: PflichtKommentarSchema,
})
export type StornierungErteilen = z.infer<typeof StornierungErteilenSchema>
