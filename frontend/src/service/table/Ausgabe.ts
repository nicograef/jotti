import { z } from 'zod'

import { DateStringSchema } from '../schemas'
import { PositionSchema } from './Bestellung'

export const AusgabeSchema = z.object({
  art: z.literal('ausgabe'),
  id: z.uuid(),
  userId: z.number().int().min(1),
  // Eingefrorener Name des Akteurs, beschriftet den Historien-Eintrag.
  userName: z.string().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  kommentar: z.string().max(100),
  ausgegebenAm: DateStringSchema,
})
export type Ausgabe = z.infer<typeof AusgabeSchema>
