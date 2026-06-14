import { z } from 'zod'

import { DateStringSchema } from '../schemas'
import { PositionRefSchema, PositionSchema } from './Bestellung'

export const StornierungSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtStornierungCents: z.number().int().min(0),
  kommentar: z.string().min(3).max(100),
  storniertAm: DateStringSchema,
})
export type Stornierung = z.infer<typeof StornierungSchema>

export const StornierungErteilenSchema = z.object({
  tischId: z.number().int().min(1),
  positionen: PositionRefSchema.array().min(1),
  kommentar: z.string().min(3).max(100),
})
export type StornierungErteilen = z.infer<typeof StornierungErteilenSchema>
