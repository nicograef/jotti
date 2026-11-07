import { z } from 'zod'

import { DateStringSchema } from '@/lib/utils'

import {
  KommentarSchema,
  PfandVerrechnungSchema,
  PositionRefSchema,
  TischIdSchema,
} from '../schemas'
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
  kommentar: KommentarSchema,
  kassiertAm: DateStringSchema,
})
export type Zahlung = z.infer<typeof ZahlungSchema>

export const ZahlungKassierenSchema = z.object({
  zahlungId: z.uuid(),
  tischId: TischIdSchema,
  positionen: PositionRefSchema.array().min(1),
  kommentar: KommentarSchema,
  pfand: PfandVerrechnungSchema.optional(),
})
export type ZahlungKassieren = z.infer<typeof ZahlungKassierenSchema>
