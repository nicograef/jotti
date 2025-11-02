import { z } from "zod"

export const UserRole = {
  ADMIN: "admin",
  SERVICE: "service",
} as const
export type UserRole = (typeof UserRole)[keyof typeof UserRole]

export const UserSchema = z.object({
  id: z.int().min(1),
  name: z.string().min(5).max(50),
  username: z
    .string()
    .min(3)
    .max(20)
    .regex(/^[a-z0-9]+$/, {
      message: "Username can only contain lowercase letters and numbers",
    }),
  role: z.enum(UserRole),
  createdAt: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: "Invalid date format",
  }),
  gesperrt: z.boolean(),
})

export type User = z.infer<typeof UserSchema>
