import { z } from 'zod'

import { EntityStatusSchema } from '@/lib/entityStatus'
import { UsernameSchema, UserRoleSchema } from '@/lib/identity'
import { createNameSchema } from '@/lib/nameSchema'
import { DateStringSchema } from '@/lib/utils'

export const UserIdSchema = z.number().int().min(1)
const NameSchema = createNameSchema(50)

export const UserSchema = z.object({
  id: UserIdSchema,
  name: NameSchema,
  username: UsernameSchema,
  role: UserRoleSchema,
  createdAt: DateStringSchema,
  status: EntityStatusSchema,
  updatedAt: DateStringSchema,
})
export type User = z.infer<typeof UserSchema>
