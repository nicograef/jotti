import { z } from 'zod'

// Credential rules shared by auth and user management, each mirroring its zog counterpart in domain/user.
// The trim matters: the backend stores the trimmed value, so an untrimmed pasted credential would not match.

export const UsernameSchema = z
  .string()
  .trim()
  .min(3, { message: 'Benutzername muss mindestens 3 Zeichen lang sein.' })
  .max(20, { message: 'Benutzername darf maximal 20 Zeichen lang sein.' })
  .regex(/^[a-z0-9]+$/, {
    message: 'Benutzername darf nur aus Kleinbuchstaben und Zahlen bestehen.',
  })

export const PasswordSchema = z
  .string()
  .trim()
  .min(6, { message: 'Passwort muss mindestens 6 Zeichen lang sein.' })
  .max(72, { message: 'Passwort darf maximal 72 Zeichen lang sein.' })

export const OnetimePasswordSchema = z
  .string()
  .trim()
  .regex(/^\d{6}$/, {
    message: 'Das Einmalpasswort besteht aus genau 6 Ziffern.',
  })

// Normalizes free input to a valid username (lower case, no spaces, umlauts spelled out, rest removed),
// so the field meets UsernameSchema instead of naming it on error.
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
