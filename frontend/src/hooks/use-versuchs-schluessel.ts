import { useRef } from 'react'

// Schlüssel eines Absendeversuchs: ein Retry mit gleichem Inhalt behält
// ihn, damit eine verlorene Antwort nicht doppelt bucht; anderer Inhalt bekommt
// einen neuen. Ein Vorgang mit gehobener Auswahl nutzt useVorgang.
export function useVersuchsSchluessel(): (inhalt: unknown) => string {
  const versuch = useRef<{ inhalt: string; schluessel: string } | null>(null)
  return (inhalt) => {
    const json = JSON.stringify(inhalt)
    if (versuch.current?.inhalt !== json) {
      versuch.current = { inhalt: json, schluessel: crypto.randomUUID() }
    }
    return versuch.current.schluessel
  }
}
