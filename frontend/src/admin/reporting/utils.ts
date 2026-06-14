export function pct(part: number, total: number): number {
  return total > 0 ? Math.round((part / total) * 100) : 0
}

export function formatLocalTime(utcString: string): string {
  return new Date(utcString).toLocaleString('de-DE')
}
