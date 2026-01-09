import { z } from 'zod'

import { OrderProductSchema } from './Order'

export const CancelationSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tableId: z.number().int().min(1),
  products: OrderProductSchema.array().min(1),
  totalCancelationCents: z.number().int().min(0),
  canceledAt: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Cancelation = z.infer<typeof CancelationSchema>

export const CancelProductsSchema = z.object({
  tableId: z.number().int().min(1),
  products: OrderProductSchema.array().min(1),
})
export type CancelProducts = z.infer<typeof CancelProductsSchema>
