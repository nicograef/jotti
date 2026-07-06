import { z } from 'zod'

// Shared credential/identity validation rules, used by both auth (login,
// set-password) and user management (create/reset). Single source of truth so
// the rules can never drift between the two areas. Bounds mirror the backend
// (domain/user): username 3–20 lowercase-alphanumeric, password 6–72,
// one-time password exactly 6 digits.

export const UsernameSchema = z
  .string()
  .min(3, { message: 'Benutzername muss mindestens 3 Zeichen lang sein.' })
  .max(20, { message: 'Benutzername darf maximal 20 Zeichen lang sein.' })
  .regex(/^[a-z0-9]+$/, {
    message: 'Benutzername darf nur aus Kleinbuchstaben und Zahlen bestehen.',
  })

export const PasswordSchema = z
  .string()
  .min(6, { message: 'Passwort muss mindestens 6 Zeichen lang sein.' })
  .max(72, { message: 'Passwort darf maximal 72 Zeichen lang sein.' })

export const OnetimePasswordSchema = z.string().regex(/^\d{6}$/, {
  message: 'Das Einmalpasswort besteht aus genau 6 Ziffern.',
})
