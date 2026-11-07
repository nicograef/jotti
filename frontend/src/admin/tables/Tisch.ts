import { z } from 'zod'

import { EntityStatusSchema } from '@/lib/entityStatus'
import { createNameSchema } from '@/lib/nameSchema'

export const TischIdSchema = z.number().int().min(1)
const TischNameSchema = createNameSchema(100)

export const TischSchema = z.object({
  id: TischIdSchema,
  name: TischNameSchema,
  status: EntityStatusSchema,
  // Offener Saldo in der aktuell offenen Kassensitzung (Cent; 0 ohne Saldo oder
  // ohne offene Sitzung). Reine Backend-Projektion.
  saldoCents: z.number().int().min(0),
})
export type Tisch = z.infer<typeof TischSchema>

// Sammelgruppe für Tische ohne abschließende Zahl im Namen.
export const WEITERE_GRUPPE = 'Weitere'
