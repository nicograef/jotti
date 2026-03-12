import { z } from 'zod'

import { PositionSchema } from './Bestellung'

export const StornierungSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtStornierungCents: z.number().int().min(0),
  kommentar: z.string().max(100),
  storniertAm: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Stornierung = z.infer<typeof StornierungSchema>

export const ProdukteStornierenSchema = z.object({
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type ProdukteStornieren = z.infer<typeof ProdukteStornierenSchema>
