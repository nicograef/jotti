import { z } from 'zod'

import { createNameSchema } from '@/lib/nameSchema'
import { DateStringSchema } from '@/lib/utils'

export const TischIdSchema = z.number().int().min(1)
const TischNameSchema = createNameSchema(100)

export const TischStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const
export type TischStatus = (typeof TischStatus)[keyof typeof TischStatus]
const TischStatusSchema = z.enum(TischStatus)

export const TischSchema = z.object({
  id: TischIdSchema,
  name: TischNameSchema,
  status: TischStatusSchema,
  // Offener Saldo in der aktuell offenen Kassensitzung (Cent; 0 ohne Saldo oder
  // ohne offene Sitzung). Reine Backend-Projektion.
  saldoCents: z.number().int().min(0),
  createdAt: DateStringSchema,
  updatedAt: DateStringSchema,
})
export type Tisch = z.infer<typeof TischSchema>

// Sammelgruppe für Tische ohne abschließende Zahl im Namen.
export const WEITERE_GRUPPE = 'Weitere'
