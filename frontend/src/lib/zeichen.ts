import { z } from 'zod'

// Zählt Unicode-Codepoints wie das Backend (utf8.RuneCountInString), nicht UTF-16-Einheiten.
export function zeichenAnzahl(text: string): number {
  return [...text].length
}

// Kürzt auf höchstens max Zeichen, ohne einen Codepoint (Emoji) zu zerschneiden.
export function kuerzeAufZeichen(text: string, max: number): string {
  return [...text].slice(0, max).join('')
}

export function maxZeichen(schema: z.ZodString, max: number, message?: string) {
  return schema.refine((text) => zeichenAnzahl(text) <= max, { message })
}
