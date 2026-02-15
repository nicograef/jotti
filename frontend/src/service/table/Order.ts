import { z } from 'zod'

export const LineItemSchema = z.object({
  id: z.number().int().min(1),
  name: z.string().min(1).max(100),
  priceCents: z.number().int().min(0),
  quantity: z.number().int().min(1),
})
export type LineItem = z.infer<typeof LineItemSchema>

export const PlaceOrderSchema = z.object({
  tableId: z.number().int().min(1),
  variants: LineItemSchema.array().min(1),
  comment: z.string().max(100),
})
export type PlaceOrder = z.infer<typeof PlaceOrderSchema>

export const OrderSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tableId: z.number().int().min(1),
  variants: LineItemSchema.array().min(1),
  totalPriceCents: z.number().int().min(0),
  comment: z.string().max(100),
  placedAt: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Order = z.infer<typeof OrderSchema>
