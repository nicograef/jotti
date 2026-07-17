import { z } from 'zod'

export { DateStringSchema } from '@/lib/utils'

export const SteuersatzSchema = z.enum([
  'regel',
  'ermaessigt',
  'befreit',
  'kombi',
])

// Referenz auf eine (Teil-)Menge einer Position, geteilt von Tisch- und
// Direktverkauf-Vorgängen (Kassieren, Stornieren, Umbuchen).
export const PositionRefSchema = z.object({
  positionId: z.uuid(),
  menge: z.number().int().min(1),
})
export type PositionRef = z.infer<typeof PositionRefSchema>
