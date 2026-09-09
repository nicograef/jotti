import { z } from 'zod'

export { DateStringSchema } from '@/lib/utils'

// Referenz auf eine (Teil-)Menge einer Position, geteilt von Tisch- und
// Direktverkauf-Vorgängen (Kassieren, Stornieren, Umbuchen).
export const PositionRefSchema = z.object({
  positionId: z.uuid(),
  menge: z.number().int().min(1),
})
export type PositionRef = z.infer<typeof PositionRefSchema>
