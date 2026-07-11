// Admin-Auswertungen zeigen den eingefrorenen Username, ergaenzt um den live
// aufgeloesten Klarnamen: "username (Klarname)". Fehlt der Klarname, nur Username.
export function formatBediener(userName: string, name: string): string {
  return name ? `${userName} (${name})` : userName
}

// formatUhrzeit formatiert einen Zeitpunkt als HH:MM in lokaler Zeit — die
// einzige Stelle, die diese Uhrzeit-Optionen festlegt.
function formatUhrzeit(date: Date): string {
  return date.toLocaleTimeString('de-DE', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

// Storno-Zeitstempel innerhalb einer Kassensitzung (ein Tag) als HH:MM in
// lokaler Zeit; das Datum ergibt sich aus der Kassensitzung.
export function formatLocalTime(utcString: string): string {
  return formatUhrzeit(new Date(utcString))
}

// formatStand formatiert den Aktualitäts-Zeitpunkt (ms seit Epoch, aus React
// Query dataUpdatedAt) als HH:MM für den "Stand HH:MM"-Hinweis des Live-Dashboards.
export function formatStand(dataUpdatedAt: number): string {
  return formatUhrzeit(new Date(dataUpdatedAt))
}

export function formatDatum(datum: string): string {
  return new Date(datum).toLocaleDateString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    timeZone: 'UTC',
  })
}
