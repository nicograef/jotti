import { z } from 'zod'

import { UsernameSchema } from '@/lib/identity'
import { DateStringSchema } from '@/lib/utils'

export const UserRole = {
  ADMIN: 'admin',
  SERVICELEITUNG: 'serviceleitung',
  SERVICE: 'service',
} as const
export type UserRole = (typeof UserRole)[keyof typeof UserRole]

export function toUsername(name: string) {
  return name
    .toLowerCase()
    .replace(/\s+/g, '')
    .replace(/ä/g, 'ae')
    .replace(/ö/g, 'oe')
    .replace(/ü/g, 'ue')
    .replace(/ß/g, 'ss')
    .replace(/[^a-z0-9]/g, '')
}

export const UserStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const
export type UserStatus = (typeof UserStatus)[keyof typeof UserStatus]

export const UserIdSchema = z.number().int().min(1)
const NameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(50, { message: 'Der Name ist zu lang.' })
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
