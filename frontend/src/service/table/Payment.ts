import { z } from 'zod'

import { OrderVariantSchema } from './Order'

export const PaymentSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tableId: z.number().int().min(1),
  variants: OrderVariantSchema.array().min(1),
  totalPaymentCents: z.number().int().min(0),
  comment: z.string().max(100),
  registeredAt: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Payment = z.infer<typeof PaymentSchema>

export const RegisterPaymentSchema = z.object({
  tableId: z.number().int().min(1),
  variants: OrderVariantSchema.array().min(1),
  comment: z.string().max(100),
})
export type RegisterPayment = z.infer<typeof RegisterPaymentSchema>
