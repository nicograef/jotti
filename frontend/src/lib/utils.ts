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
 * Formats a price in cents as a Euro amount followed by the € sign, joined by a
 * non-breaking space (U+00A0) so amount and sign never wrap apart.
 * Example: 1250 → "12,50 €"
 */
export function formatEuro(cents: number): string {
  return `${formatCents(cents)}\u00A0€`
}

/**
 * Label of the "Alle auswählen" button in Kassieren and Umbuchung: correct
 * grammatical number (singular "1 Position", plural "N Positionen") plus the
 * selection sum. With exactly one position the plural-implying "Alle" is dropped.
 */
export function formatAlleAuswaehlenLabel(
  anzahl: number,
  summeCents: number,
): string {
  const auswahl =
    anzahl === 1
      ? '1 Position auswählen'
      : `Alle ${anzahl.toString()} Positionen auswählen`
  return `${auswahl} · ${formatEuro(summeCents)}`
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
 * Formats a timestamp relative to now, for scanning history lists: "gerade eben"
 * (< 1 min), "vor X min" (< 60 min), "vor X Std" (< 6 h), otherwise the absolute
 * clock time "18:42" (same day) or "11.7., 18:42" (earlier day). No live ticker —
 * the value only changes on re-render/refetch, which is accepted. The full
 * timestamp stays available in the detail drawer.
 */
export function formatRelativeTime(
  date: string,
  now: Date = new Date(),
): string {
  const then = new Date(date)
  const diffMs = now.getTime() - then.getTime()
  const diffMin = Math.floor(diffMs / 60_000)

  if (diffMin < 1) return 'gerade eben'
  if (diffMin < 60) return `vor ${diffMin.toString()} min`

  const diffStd = Math.floor(diffMin / 60)
  if (diffStd < 6) return `vor ${diffStd.toString()} Std`

  const uhrzeit = then.toLocaleTimeString('de-DE', {
    hour: '2-digit',
    minute: '2-digit',
  })
  const gleicherTag =
    then.getFullYear() === now.getFullYear() &&
    then.getMonth() === now.getMonth() &&
    then.getDate() === now.getDate()
  if (gleicherTag) return uhrzeit

  const datum = then.toLocaleDateString('de-DE', {
    day: 'numeric',
    month: 'numeric',
  })
  return `${datum}, ${uhrzeit}`
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
