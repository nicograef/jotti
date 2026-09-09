import { z } from 'zod'

import { UsernameSchema } from '@/lib/identity'
import { createNameSchema } from '@/lib/nameSchema'
import { DateStringSchema } from '@/lib/utils'

export const UserRole = {
  ADMIN: 'admin',
  SERVICELEITUNG: 'serviceleitung',
  SERVICE: 'service',
} as const
export type UserRole = (typeof UserRole)[keyof typeof UserRole]

export const UserStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const
export type UserStatus = (typeof UserStatus)[keyof typeof UserStatus]

export const UserIdSchema = z.number().int().min(1)
const NameSchema = createNameSchema(50)
const RoleSchema = z.enum(UserRole)
const UserStatusSchema = z.enum(UserStatus)

export const UserSchema = z.object({
  id: UserIdSchema,
  name: NameSchema,
  username: UsernameSchema,
  role: RoleSchema,
  createdAt: DateStringSchema,
  status: UserStatusSchema,
  updatedAt: DateStringSchema,
})
export type User = z.infer<typeof UserSchema>
