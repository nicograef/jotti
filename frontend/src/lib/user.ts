import { z } from "zod"

export const UserRole = {
  ADMIN: "admin",
  SERVICE: "service",
} as const
export type UserRole = (typeof UserRole)[keyof typeof UserRole]

export function toUsername(name: string) {
  return name
    .toLowerCase()
    .replace(/\s+/g, "")
    .replace(/ä/g, "ae")
    .replace(/ö/g, "oe")
    .replace(/ü/g, "ue")
    .replace(/ß/g, "ss")
    .replace(/[^a-z0-9]/g, "")
}

const UserIdSchema = z.number().min(1)
const NameSchema = z.string().min(5).max(50)
const UsernameSchema = z
  .string()
  .min(3)
  .max(20)
  .regex(/^[a-z0-9]+$/, {
    message: "Username can only contain lowercase letters and numbers",
  })
const RoleSchema = z.enum(UserRole)
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: "Invalid date format",
})
const PasswordSchema = z.string().min(6).max(20)
const OnetimePasswordSchema = z.string().regex(/^\d{6}$/, {
  message: "Onetime password must be exactly 6 digits",
})

export const UserSchema = z.object({
  id: UserIdSchema,
  name: NameSchema,
  username: UsernameSchema,
  role: RoleSchema,
  createdAt: DateStringSchema,
  locked: z.boolean(),
})
export type User = z.infer<typeof UserSchema>

export const SetPasswordRequestSchema = z.object({
  username: UsernameSchema,
  password: PasswordSchema,
  onetimePassword: OnetimePasswordSchema,
})

export const CreateUserRequestSchema = z.object({
  name: NameSchema,
  username: UsernameSchema,
  role: RoleSchema,
})
export type CreateUserRequest = z.infer<typeof CreateUserRequestSchema>

export const CreateUserResponseSchema = z.object({
  user: UserSchema,
  onetimePassword: OnetimePasswordSchema,
})
export type CreateUserResponse = z.infer<typeof CreateUserResponseSchema>

export const GetUsersResponseSchema = z.object({
  users: UserSchema.array(),
})
