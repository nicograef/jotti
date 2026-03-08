import { z } from 'zod'

export const TischIdSchema = z.number().int().min(1)
const TischNameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(100, { message: 'Der Name ist zu lang.' })
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: 'Ungültiges Datumsformat',
})

export const TischStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const
export type TischStatus = (typeof TischStatus)[keyof typeof TischStatus]
const TischStatusSchema = z.enum(TischStatus)

export const TischSchema = z.object({
  id: TischIdSchema,
  name: TischNameSchema,
  status: TischStatusSchema,
  createdAt: DateStringSchema,
})
export type Tisch = z.infer<typeof TischSchema>
