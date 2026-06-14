export function pct(part: number, total: number): number {
  return total > 0 ? Math.round((part / total) * 100) : 0
}

export function formatLocalTime(utcString: string): string {
  return new Date(utcString).toLocaleString('de-DE')
}

export function formatDatum(datum: string): string {
  return new Date(datum).toLocaleDateString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    timeZone: 'UTC',
  })
}
