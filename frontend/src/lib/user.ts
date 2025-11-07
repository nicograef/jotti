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
const NameSchema = z
  .string()
  .min(5, { message: "Das sieht nicht nach einem echten Namen aus." })
  .max(50, { message: "Der Name ist zu lang." })
const UsernameSchema = z
  .string()
  .min(3, { message: "Benutzername muss mindestens 3 Zeichen lang sein." })
  .max(20, { message: "Benutzername darf maximal 20 Zeichen lang sein." })
  .regex(/^[a-z0-9]+$/, {
    message: "Benutzername darf nur aus Kleinbuchstaben und Zahlen bestehen.",
  })
const RoleSchema = z.enum(UserRole)
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: "Ungültiges Datumsformat",
})
const PasswordSchema = z.string().min(6, { message: "Passwort muss mindestens 6 Zeichen lang sein." }).max(20, { message: "Passwort darf maximal 20 Zeichen lang sein." })
const OnetimePasswordSchema = z.string().regex(/^\d{6}$/, {
  message: "Das Einmalpasswort muss genau 6 Ziffern enthalten.",
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

export const UpdateUserRequestSchema = UserSchema.pick({
  id: true,
  name: true,
  username: true,
  role: true,
  locked: true,
})
export type UpdateUserRequest = z.infer<typeof UpdateUserRequestSchema>

export const UpdateUserResponseSchema = z.object({
  user: UserSchema,
})
export type UpdateUserResponse = z.infer<typeof UpdateUserResponseSchema>

export const GetUsersResponseSchema = z.object({
  users: UserSchema.array(),
})
