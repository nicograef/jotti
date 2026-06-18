export function pct(part: number, total: number): number {
  return total > 0 ? Math.round((part / total) * 100) : 0
}

// Admin-Auswertungen zeigen den eingefrorenen Username, ergaenzt um den live
// aufgeloesten Klarnamen: "username (Klarname)". Fehlt der Klarname, nur Username.
export function formatBediener(userName: string, name: string): string {
  return name ? `${userName} (${name})` : userName
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

// formatOffeneArbeit beschreibt die offene eigene Arbeit an einem Tisch: noch
// auszugebende (ausstehende) und noch zu kassierende (unbezahlte) Positionen.
// Null-Anteile werden weggelassen.
export function formatOffeneArbeit(tisch: {
  anzahlAusstehend: number
  anzahlUnbezahlt: number
}): string {
  const parts: string[] = []
  if (tisch.anzahlAusstehend > 0) {
    parts.push(`${tisch.anzahlAusstehend.toString()} auszugeben`)
  }
  if (tisch.anzahlUnbezahlt > 0) {
    parts.push(`${tisch.anzahlUnbezahlt.toString()} zu kassieren`)
  }
  return parts.join(' · ')
}
