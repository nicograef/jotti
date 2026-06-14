import { z } from 'zod'

import { DateStringSchema } from '../schemas'

export const AuszahlungSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tischId: z.number().int().min(1),
  betragCents: z.number().int().min(1),
  kommentar: z.string().min(3).max(100),
  geleistetAm: DateStringSchema,
})
export type Auszahlung = z.infer<typeof AuszahlungSchema>

export const AuszahlungLeistenSchema = z.object({
  tischId: z.number().int().min(1),
  betragCents: z.number().int().min(1),
  kommentar: z.string().min(3).max(100),
})
export type AuszahlungLeisten = z.infer<typeof AuszahlungLeistenSchema>
