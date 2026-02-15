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
