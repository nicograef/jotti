import { z } from 'zod'

// Shared credential rules for auth (login, set-password) and user management
// (create/reset), each mirroring its zog counterpart in the backend
// (domain/user). The trim matters because the backend stores the trimmed value:
// without it a pasted credential with spaces no longer matches what was stored.

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

// Normalisiert eine freie Eingabe zu einem gültigen Benutzernamen: klein
// geschrieben, ohne Leerzeichen, Umlaute ausgeschrieben, alles Übrige entfernt.
// Damit trifft das Eingabefeld die Regel von UsernameSchema, statt sie erst im
// Fehlerfall zu nennen.
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
