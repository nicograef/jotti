import { z } from 'zod'

export const OrderProductSchema = z.object({
  id: z.number().int().min(1),
  name: z.string().min(1).max(100),
  netPriceCents: z.number().int().min(0),
  quantity: z.number().int().min(1),
})
export type OrderProduct = z.infer<typeof OrderProductSchema>

export const PlaceOrderSchema = z.object({
  tableId: z.number().int().min(1),
  products: OrderProductSchema.array().min(1),
})
export type PlaceOrder = z.infer<typeof PlaceOrderSchema>

export const OrderSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tableId: z.number().int().min(1),
  products: OrderProductSchema.array().min(1),
  totalNetPriceCents: z.number().int().min(0),
  placedAt: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Order = z.infer<typeof OrderSchema>
