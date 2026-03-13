import { z } from 'zod'

import { PositionRefSchema } from './Bestellung'

export const LieferungSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionRefSchema.array().min(1),
  kommentar: z.string().max(100),
  geliefertAm: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Lieferung = z.infer<typeof LieferungSchema>

export const ProdukteLiefernSchema = z.object({
  tischId: z.number().int().min(1),
  positionen: PositionRefSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type ProdukteLiefern = z.infer<typeof ProdukteLiefernSchema>
