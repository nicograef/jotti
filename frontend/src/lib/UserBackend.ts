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
  DELETED: 'deleted',
} as const
export type UserStatus = (typeof UserStatus)[keyof typeof UserStatus]

const UserIdSchema = z.number().int().min(1)
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
const OnetimePasswordSchema = z.string().regex(/^\d{6}$/, {
  message: 'Das Einmalpasswort muss genau 6 Ziffern enthalten.',
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

export const CreateUserRequestSchema = z.object({
  name: NameSchema,
  username: UsernameSchema,
  role: RoleSchema,
})
const CreateUserResponseSchema = z.object({
  user: UserSchema,
  onetimePassword: OnetimePasswordSchema,
})

const UpdateUserRequestSchema = UserSchema.pick({
  id: true,
  name: true,
  username: true,
  role: true,
})
const UpdateUserResponseSchema = z.object({
  user: UserSchema,
})

const GetUsersResponseSchema = z.object({
  users: UserSchema.array(),
})

interface Backend {
  post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse>
}

export class UserBackend {
  private readonly backend: Backend

  constructor(backend: Backend) {
    this.backend = backend
  }

  public async createUser(
    name: string,
    username: string,
    role: UserRole,
  ): Promise<{ user: User; onetimePassword: string }> {
    const body = CreateUserRequestSchema.parse({ name, username, role })
    const { user, onetimePassword } = await this.backend.post(
      'admin/create-user',
      body,
      CreateUserResponseSchema,
    )
    return { user, onetimePassword }
  }

  public async updateUser(
    updatedUser: z.infer<typeof UpdateUserRequestSchema>,
  ): Promise<User> {
    const body = UpdateUserRequestSchema.parse(updatedUser)
    const { user } = await this.backend.post(
      'admin/update-user',
      body,
      UpdateUserResponseSchema,
    )
    return user
  }

  public async getUsers(): Promise<User[]> {
    const { users } = await this.backend.post(
      'admin/get-users',
      {},
      GetUsersResponseSchema,
    )
    return users
  }

  public async activateUser(id: number): Promise<void> {
    const body = z.object({ id: UserIdSchema }).parse({ id })
    await this.backend.post('admin/activate-user', body)
  }

  public async deactivateUser(id: number): Promise<void> {
    const body = z.object({ id: UserIdSchema }).parse({ id })
    await this.backend.post('admin/deactivate-user', body)
  }
}
