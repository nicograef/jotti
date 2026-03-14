import { z } from 'zod'

import { PositionSchema } from './Bestellung'

export const TischIdSchema = z.number().int().min(1)
const TischNameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(100, { message: 'Der Name ist zu lang.' })

export const TischSchema = z.object({
  id: TischIdSchema,
  name: TischNameSchema,
})
export type Tisch = z.infer<typeof TischSchema>

export const TischStateSchema = z.object({
  tischId: TischIdSchema,
  tischName: z.string(),
  saldoCents: z.number().int(),
  unbezahltePositionen: z.array(PositionSchema),
  ausstehendePositionen: z.array(PositionSchema),
  gesamtZahlungenCents: z.number().int(),
})
export type TischState = z.infer<typeof TischStateSchema>
