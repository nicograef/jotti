import { z } from 'zod'

import { OrderVariantSchema } from './Order'

export const DeliverySchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tableId: z.number().int().min(1),
  variants: OrderVariantSchema.array().min(1),
  comment: z.string().max(100),
  deliveredAt: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Delivery = z.infer<typeof DeliverySchema>

export const DeliverVariantsSchema = z.object({
  tableId: z.number().int().min(1),
  variants: OrderVariantSchema.array().min(1),
  comment: z.string().max(100),
})
export type DeliverVariants = z.infer<typeof DeliverVariantsSchema>
