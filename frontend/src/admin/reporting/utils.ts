// Admin-Auswertungen zeigen den eingefrorenen Username, ergaenzt um den live
// aufgeloesten Klarnamen: "username (Klarname)". Fehlt der Klarname, nur Username.
export function formatBediener(userName: string, name: string): string {
  return name ? `${userName} (${name})` : userName
}

// Storno-Zeitstempel innerhalb einer Kassensitzung (ein Tag) als HH:MM in
// lokaler Zeit; das Datum ergibt sich aus der Kassensitzung.
export function formatLocalTime(utcString: string): string {
  return new Date(utcString).toLocaleTimeString('de-DE', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatDatum(datum: string): string {
  return new Date(datum).toLocaleDateString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    timeZone: 'UTC',
  })
}

// formatOffeneArbeit beschreibt die offene eigene Arbeit an einem Tisch: die noch
// zu kassierenden (unbezahlten) Positionen.
export function formatOffeneArbeit(tisch: { anzahlUnbezahlt: number }): string {
  return `${tisch.anzahlUnbezahlt.toString()} zu kassieren`
}
