import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import {
  OnetimePasswordSchema,
  PasswordSchema,
  UsernameSchema,
} from '@/lib/identity'

export const LoginSchema = z.object({
  username: UsernameSchema,
  password: PasswordSchema,
})
const LoginResponseSchema = z.object({
  token: z.string().min(10),
})

export const SetPasswordSchema = z.object({
  username: UsernameSchema,
  password: PasswordSchema,
  onetimePassword: OnetimePasswordSchema,
})

export class AuthBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  /** Sends a login request with the given username and password and returns the JWT token from the backend. */
  public async login(username: string, password: string): Promise<string> {
    const body = LoginSchema.parse({ username, password })
    const { token } = await this.backend.post(
      'auth/login',
      body,
      LoginResponseSchema,
    )
    return token
  }

  /** Sets the initial password for an account, authorized by its one-time password. */
  public async setPassword(
    username: string,
    password: string,
    onetimePassword: string,
  ): Promise<void> {
    const body = SetPasswordSchema.parse({
      username,
      password,
      onetimePassword,
    })
    await this.backend.post('auth/set-password', body)
  }
}
