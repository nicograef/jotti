import { z } from 'zod'

import { OrderVariantSchema } from './Order'

export const CancelationSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tableId: z.number().int().min(1),
  variants: OrderVariantSchema.array().min(1),
  totalCancelationCents: z.number().int().min(0),
  comment: z.string().max(100),
  canceledAt: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Cancelation = z.infer<typeof CancelationSchema>

export const CancelVariantsSchema = z.object({
  tableId: z.number().int().min(1),
  variants: OrderVariantSchema.array().min(1),
  comment: z.string().max(100),
})
export type CancelVariants = z.infer<typeof CancelVariantsSchema>
