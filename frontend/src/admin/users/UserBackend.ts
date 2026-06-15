import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import { OnetimePasswordSchema } from '@/lib/identity'

import { type User, UserIdSchema, UserRole, UserSchema } from './User'

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

export class UserBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
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

  public async deleteUser(id: number): Promise<void> {
    const body = UserSchema.pick({ id: true }).parse({ id })
    await this.backend.post('admin/delete-user', body)
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
