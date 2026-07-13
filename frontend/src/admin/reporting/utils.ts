// Admin-Auswertungen zeigen den eingefrorenen Username, ergänzt um den live
// aufgelösten Klarnamen: "username (Klarname)". Fehlt der Klarname, nur Username.
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

// formatDatumKurz gibt Wochentag und Tag/Monat für die Sitzungslisten-Karten
// aus ("Fr, 05.07."). Der Kalendertag der Kassensitzung ist UTC-normiert.
export function formatDatumKurz(datum: string): string {
  return new Date(datum).toLocaleDateString('de-DE', {
    weekday: 'short',
    day: '2-digit',
    month: '2-digit',
    timeZone: 'UTC',
  })
}

// formatDatumLang gibt Wochentag und vollständiges Datum für den Berichtskopf
// aus ("Fr, 05.07.2026").
export function formatDatumLang(datum: string): string {
  return new Date(datum).toLocaleDateString('de-DE', {
    weekday: 'short',
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    timeZone: 'UTC',
  })
}
