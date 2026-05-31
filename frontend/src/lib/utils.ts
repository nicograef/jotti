import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Formats a price in cents as a Euro string with comma decimal separator.
 * Example: 1250 → "12,50"
 */
export function formatCents(cents: number): string {
  return (cents / 100).toFixed(2).replace('.', ',')
}

/**
 * Parses a Euro string (with comma or dot separator) to cents.
 * Example: "12,50" → 1250, "12.50" → 1250, invalid → 0
 */
export function parseCents(euroInput: string): number {
  const parsed = parseFloat(euroInput.replace(',', '.'))
  return isNaN(parsed) ? 0 : Math.round(parsed * 100)
}
