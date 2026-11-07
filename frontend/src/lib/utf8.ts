import { z } from 'zod'

const encoder = new TextEncoder()

export function utf8Bytes(text: string): number {
  return encoder.encode(text).length
}

// Kürzt auf höchstens maxBytes UTF-8-Bytes, ohne einen Codepoint (Umlaut, Emoji) zu zerschneiden.
export function kuerzeAufUtf8Bytes(text: string, maxBytes: number): string {
  if (utf8Bytes(text) <= maxBytes) {
    return text
  }
  let bytes = 0
  let ende = 0
  for (const zeichen of text) {
    bytes += utf8Bytes(zeichen)
    if (bytes > maxBytes) {
      break
    }
    ende += zeichen.length
  }
  return text.slice(0, ende)
}

// zogs Max zählt UTF-8-Bytes, nicht Zeichen: ein String-Limit, das das Backend
// annimmt, muss hier ebenfalls in Bytes gezählt werden.
export function maxUtf8Bytes(
  schema: z.ZodString,
  maxBytes: number,
  message?: string,
) {
  return schema.refine((text) => utf8Bytes(text) <= maxBytes, { message })
}
