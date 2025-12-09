import { z } from 'zod'

import { type User, UserIdSchema, UserRole, UserSchema } from './User'

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

  public async getAllUsers(): Promise<User[]> {
    const { users } = await this.backend.post(
      'admin/get-all-users',
      {},
      z.object({ users: UserSchema.array() }),
    )
    return users
  }

  public async createUser(
    name: string,
    username: string,
    role: UserRole,
  ): Promise<{ id: number; onetimePassword: string }> {
    const body = CreateUserSchema.parse({ name, username, role })
    const { id, onetimePassword } = await this.backend.post(
      'admin/create-user',
      body,
      z.object({ id: UserIdSchema, onetimePassword: OnetimePasswordSchema }),
    )
    return { id, onetimePassword }
  }

  public async updateUser(
    updatedUser: z.infer<typeof UpdateUserSchema>,
  ): Promise<void> {
    const body = UpdateUserSchema.parse(updatedUser)
    await this.backend.post('admin/update-user', body)
  }

  public async activateUser(id: number): Promise<void> {
    const body = UserSchema.pick({ id: true }).parse({ id })
    await this.backend.post('admin/activate-user', body)
  }

  public async deactivateUser(id: number): Promise<void> {
    const body = UserSchema.pick({ id: true }).parse({ id })
    await this.backend.post('admin/deactivate-user', body)
  }

  public async resetPassword(id: number): Promise<string> {
    const body = UserSchema.pick({ id: true }).parse({ id })
    const { onetimePassword } = await this.backend.post(
      'admin/reset-password',
      body,
      z.object({ onetimePassword: OnetimePasswordSchema }),
    )
    return onetimePassword
  }
}
