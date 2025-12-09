import { z } from 'zod'

const TableIdSchema = z.number().int().min(1)
const TableNameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(30, { message: 'Der Name ist zu lang.' })

export const TableSchema = z.object({
  id: TableIdSchema,
  name: TableNameSchema,
})
export type Table = z.infer<typeof TableSchema>
