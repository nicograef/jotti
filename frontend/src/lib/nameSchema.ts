import { z } from 'zod'

// Single source of name messages and minimum length; only the maximum differs (100 for Tisch, Produkt, Variante;
// 50 for a Benutzer's name, not the username in lib/identity.ts). `.trim()` mirrors the backend zog `.Trim()`,
// so a blank name fails in the form instead of as an anonymous validation_error.
export function createNameSchema(maxLength: number) {
  return z
    .string()
    .trim()
    .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
    .max(maxLength, { message: 'Der Name ist zu lang.' })
}
