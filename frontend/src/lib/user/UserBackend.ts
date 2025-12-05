import { z } from 'zod'

import { type User, UserRole, UserSchema } from './User'

const OnetimePasswordSchema = z.string().regex(/^\d{6}$/, {
  message: 'Das Einmalpasswort muss genau 6 Ziffern enthalten.',
})

export const CreateUserSchema = UserSchema.pick({
  name: true,
  username: true,
  role: true,
})

export const UpdateUserSchema = UserSchema.pick({
  id: true,
  name: true,
  username: true,
  role: true,
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
    const body = CreateUserSchema.parse({ name, username, role })
    const { user, onetimePassword } = await this.backend.post(
      'create-user',
      body,
      z.object({ user: UserSchema, onetimePassword: OnetimePasswordSchema }),
    )
    return { user, onetimePassword }
  }

  public async updateUser(
    updatedUser: z.infer<typeof UpdateUserSchema>,
  ): Promise<User> {
    const body = UpdateUserSchema.parse(updatedUser)
    const { user } = await this.backend.post(
      'update-user',
      body,
      z.object({ user: UserSchema }),
    )
    return user
  }

  public async getAllUsers(): Promise<User[]> {
    const { users } = await this.backend.post(
      'get-all-users',
      {},
      z.object({ users: UserSchema.array() }),
    )
    return users
  }

  public async activateUser(id: number): Promise<void> {
    const body = UserSchema.pick({ id: true }).parse({ id })
    await this.backend.post('activate-user', body)
  }

  public async deactivateUser(id: number): Promise<void> {
    const body = UserSchema.pick({ id: true }).parse({ id })
    await this.backend.post('deactivate-user', body)
  }
}
