import { z } from 'zod'

export const PaymentProductSchema = z.object({
  id: z.number().int().min(1),
  name: z.string().min(1).max(100),
  netPriceCents: z.number().int().min(0),
  quantity: z.number().int().min(1),
})
export type PaymentProduct = z.infer<typeof PaymentProductSchema>

export const PaymentSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tableId: z.number().int().min(1),
  products: PaymentProductSchema.array().min(1),
  totalPaymentCents: z.number().int().min(0),
  registeredAt: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Invalid date format',
  }),
})
export type Payment = z.infer<typeof PaymentSchema>

export const RegisterPaymentSchema = z.object({
  tableId: z.number().int().min(1),
  products: PaymentProductSchema.array().min(1),
})
export type RegisterPayment = z.infer<typeof RegisterPaymentSchema>
