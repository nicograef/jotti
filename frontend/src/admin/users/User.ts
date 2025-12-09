import { z } from 'zod'

export const UserRole = {
  ADMIN: 'admin',
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
  .min(5, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(50, { message: 'Der Name ist zu lang.' })
const UsernameSchema = z
  .string()
  .min(3, { message: 'Benutzername muss mindestens 3 Zeichen lang sein.' })
  .max(20, { message: 'Benutzername darf maximal 20 Zeichen lang sein.' })
  .regex(/^[a-z0-9]+$/, {
    message: 'Benutzername darf nur aus Kleinbuchstaben und Zahlen bestehen.',
  })
const RoleSchema = z.enum(UserRole)
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: 'Ungültiges Datumsformat',
})
const UserStatusSchema = z.enum(UserStatus)

export const UserSchema = z.object({
  id: UserIdSchema,
  name: NameSchema,
  username: UsernameSchema,
  role: RoleSchema,
  createdAt: DateStringSchema,
  status: UserStatusSchema,
})
export type User = z.infer<typeof UserSchema>
