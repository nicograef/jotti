import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import { z } from 'zod'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** A string that parses to a valid date. */
export const DateStringSchema = z
  .string()
  .refine((date) => !isNaN(Date.parse(date)), {
    message: 'Ungültiges Datumsformat',
  })

/**
 * Formats a price in cents as a Euro string with comma decimal separator.
 * Example: 1250 → "12,50"
 */
export function formatCents(cents: number): string {
  return (cents / 100).toFixed(2).replace('.', ',')
}

/**
 * Composes the canonical position name: product name and variant name joined by
 * a single space, trimmed at the edges. No brackets, no dedup.
 * Example: ("Pommes", "mit Ketchup") → "Pommes mit Ketchup"
 */
export function formatPositionName(
  produktName: string,
  varianteName: string,
): string {
  return `${produktName} ${varianteName}`.trim()
}

/**
 * Parses a Euro string (with comma or dot separator, at most two decimals) to
 * cents. String-based, no float arithmetic. Invalid or over-precise input
 * (more than two decimals, multiple separators) parses to 0.
 * Example: "12,50" → 1250, "12.50" → 1250, "12,505" → 0, "1,2,3" → 0
 */
export function parseCents(euroInput: string): number {
  const match = /^(-?)(\d*)(?:[,.](\d{0,2}))?$/.exec(euroInput.trim())
  if (!match) return 0
  const [, sign, euros, decimals = ''] = match
  if (euros === '' && decimals === '') return 0
  const cents = Number(euros || '0') * 100 + Number(decimals.padEnd(2, '0'))
  return sign === '-' ? -cents : cents
}
