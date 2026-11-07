import { z } from 'zod'

// Zerlegt in Unicode-Codepoints wie das Backend (utf8.RuneCountInString), nicht in Grapheme:
// Ein Emoji mit Hautfarbe zählt dort zwei Zeichen.
function codepoints(text: string): string[] {
  // oxlint-disable-next-line typescript/no-misused-spread
  return [...text]
}

export function zeichenAnzahl(text: string): number {
  return codepoints(text).length
}

// Kürzt auf höchstens max Zeichen, ohne einen Codepoint zu zerschneiden.
export function kuerzeAufZeichen(text: string, max: number): string {
  return codepoints(text).slice(0, max).join('')
}

export function maxZeichen(schema: z.ZodString, max: number, message?: string) {
  return schema.refine((text) => zeichenAnzahl(text) <= max, { message })
}
