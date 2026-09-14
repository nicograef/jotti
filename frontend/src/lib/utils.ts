import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import { z } from 'zod'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export const DateStringSchema = z
  .string()
  .refine((date) => !isNaN(Date.parse(date)), {
    message: 'Ungültiges Datumsformat',
  })

export function formatCents(cents: number): string {
  return (cents / 100).toFixed(2).replace('.', ',')
}

/** 1250 → "12,50 €", joined by NBSP (U+00A0) so it never wraps apart. */
export function formatEuro(cents: number): string {
  return `${formatCents(cents)}\u00A0€`
}

/**
 * `'alle'` — Umbuchung (every position), `'meine'` — Kassieren (only the
 * caller's own). Both with the correct grammatical number and the selection sum.
 */
export function formatAlleAuswaehlenLabel(
  anzahl: number,
  summeCents: number,
  variante: 'alle' | 'meine' = 'alle',
): string {
  let auswahl: string
  if (variante === 'meine') {
    auswahl =
      anzahl === 1
        ? 'Meine Position auswählen'
        : `Meine ${anzahl.toString()} Positionen auswählen`
  } else {
    auswahl =
      anzahl === 1
        ? '1 Position auswählen'
        : `Alle ${anzahl.toString()} Positionen auswählen`
  }
  return `${auswahl} · ${formatEuro(summeCents)}`
}

/**
 * Like {@link formatEuro} but prefixes "+" for a positive amount — for
 * difference displays where positive means surplus (Kassensturz-Differenz).
 */
export function formatEuroMitVorzeichen(cents: number): string {
  return cents > 0 ? `+${formatEuro(cents)}` : formatEuro(cents)
}

export function formatPositionName(
  produktName: string,
  varianteName: string,
): string {
  return `${produktName} ${varianteName}`.trim()
}

/**
 * Relative timestamp for history lists. No live ticker — the value only changes
 * on re-render/refetch, which is accepted; the full timestamp stays in the
 * detail drawer.
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
 * Euro string (comma or dot, at most two decimals) to cents. String-based, no
 * float arithmetic. Invalid or over-precise input parses to 0
 * ("12,505" → 0, "1,2,3" → 0).
 */
export function parseCents(euroInput: string): number {
  const match = /^(-?)(\d*)(?:[,.](\d{0,2}))?$/.exec(euroInput.trim())
  if (!match) return 0
  const [, sign, euros, decimals = ''] = match
  if (euros === '' && decimals === '') return 0
  const cents = Number(euros || '0') * 100 + Number(decimals.padEnd(2, '0'))
  return sign === '-' ? -cents : cents
}
