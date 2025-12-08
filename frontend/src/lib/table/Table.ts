import { z } from 'zod'

export const TableIdSchema = z.number().int().min(1)
const TableNameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(30, { message: 'Der Name ist zu lang.' })
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: 'Ungültiges Datumsformat',
})

export const TableStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const
export type TableStatus = (typeof TableStatus)[keyof typeof TableStatus]
const TableStatusSchema = z.enum(TableStatus)

export const TableSchema = z.object({
  id: TableIdSchema,
  name: TableNameSchema,
  status: TableStatusSchema,
  createdAt: DateStringSchema,
})
export type Table = z.infer<typeof TableSchema>

export const TablePublicSchema = TableSchema.pick({
  id: true,
  name: true,
})
export type TablePublic = z.infer<typeof TablePublicSchema>
