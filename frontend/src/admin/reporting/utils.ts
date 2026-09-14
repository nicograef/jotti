// Zeigt den eingefrorenen Username, ergänzt um den live aufgelösten Klarnamen.
export function formatServicekraft(userName: string, name: string): string {
  return name ? `${userName} (${name})` : userName
}

function formatUhrzeit(date: Date): string {
  return date.toLocaleTimeString('de-DE', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatLocalTime(utcString: string): string {
  return formatUhrzeit(new Date(utcString))
}

// dataUpdatedAt ist ms seit Epoch (React Query).
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

// Der Kalendertag der Kassensitzung ist UTC-normiert.
export function formatDatumKurz(datum: string): string {
  return new Date(datum).toLocaleDateString('de-DE', {
    weekday: 'short',
    day: '2-digit',
    month: '2-digit',
    timeZone: 'UTC',
  })
}

export function formatDatumLang(datum: string): string {
  return new Date(datum).toLocaleDateString('de-DE', {
    weekday: 'short',
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    timeZone: 'UTC',
  })
}
