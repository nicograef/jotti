import { z } from 'zod'

export { DateStringSchema } from '@/lib/utils'

export const PositionRefSchema = z.object({
  positionId: z.uuid(),
  menge: z.number().int().min(1),
})
export type PositionRef = z.infer<typeof PositionRefSchema>
