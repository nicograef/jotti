import { z } from 'zod'

const UsernameSchema = z
  .string()
  .min(3, { message: 'Benutzername muss mindestens 3 Zeichen lang sein.' })
  .max(20, { message: 'Benutzername darf maximal 20 Zeichen lang sein.' })
  .regex(/^[a-z0-9]+$/, {
    message: 'Benutzername darf nur aus Kleinbuchstaben und Zahlen bestehen.',
  })
const PasswordSchema = z
  .string()
  .min(6, { message: 'Passwort muss mindestens 6 Zeichen lang sein.' })
  .max(20, { message: 'Passwort darf maximal 20 Zeichen lang sein.' })
const OnetimePasswordSchema = z.string().regex(/^\d{6}$/, {
  message: 'Das Einmalpasswort muss genau 6 Ziffern enthalten.',
})

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

interface Backend {
  post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse>
}

export class AuthBackend {
  private readonly backend: Backend

  constructor(backend: Backend) {
    this.backend = backend
  }

  /** Sends a login request with the given username and password and returns the JWT token from the backend. */
  public async login(username: string, password: string): Promise<string> {
    const body = LoginSchema.parse({ username, password })
    const { token } = await this.backend.post(
      'login',
      body,
      LoginResponseSchema,
    )
    return token
  }

  /** Sends a login request with the given username and password and returns the JWT token from the backend. */
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
    await this.backend.post('set-password', body)
  }
}
