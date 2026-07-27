import { z } from 'zod'

import { DateStringSchema, PositionRefSchema } from '../schemas'
import { PositionSchema } from './Bestellung'

export const ZahlungSchema = z.object({
  art: z.literal('zahlung'),
  id: z.uuid(),
  userId: z.number().int().min(1),
  // Eingefrorener Name des Akteurs, beschriftet den Historien-Eintrag.
  userName: z.string().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtZahlungCents: z.number().int().min(0),
  kommentar: z.string().max(100),
  kassiertAm: DateStringSchema,
})
export type Zahlung = z.infer<typeof ZahlungSchema>

export const ZahlungKassierenSchema = z.object({
  // Client-erzeugter Schlüssel des fachlichen Vorgangs. Ein Wiederholversuch
  // nach Verbindungsabbruch trägt denselben Schlüssel und bucht kein zweites Mal.
  vorgangId: z.uuid(),
  tischId: z.number().int().min(1),
  positionen: PositionRefSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type ZahlungKassieren = z.infer<typeof ZahlungKassierenSchema>
