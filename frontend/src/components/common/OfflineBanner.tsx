import { onlineManager } from '@tanstack/react-query'
import { useSyncExternalStore } from 'react'

// Der onlineManager startet unabhängig vom Gerätezustand mit „online“ und
// wechselt erst bei den Fenster-Events. Beim Abonnieren wird er deshalb einmal
// mit navigator.onLine abgeglichen — sonst bliebe ein Gerät, das bereits
// offline gestartet ist, für react-query dauerhaft „online“. Die Funktion liegt
// auf Modulebene, damit React genau einmal abonniert und der Abgleich nicht bei
// jedem Rendern den Zustand überschreibt.
function abonniereVerbindungszustand(aufAenderung: () => void): () => void {
  const abmelden = onlineManager.subscribe(aufAenderung)
  onlineManager.setOnline(navigator.onLine)
  return abmelden
}

function istOnline(): boolean {
  return onlineManager.isOnline()
}

// OfflineBanner meldet den Verbindungsverlust dauerhaft, solange das Gerät
// offline ist. Es liegt bewusst im normalen Fluss über dem Seiteninhalt statt
// fixiert: Kopfzeile (sticky) und Fußleiste (fixed) des Service-Bereichs
// bleiben so garantiert bedienbar.
export function OfflineBanner() {
  const online = useSyncExternalStore(abonniereVerbindungszustand, istOnline)

  if (online) {
    return null
  }

  return (
    <div
      role="status"
      className="bg-destructive px-4 py-1.5 text-center text-sm font-medium text-destructive-solid-foreground"
    >
      Keine Verbindung — Änderungen sind gerade nicht möglich
    </div>
  )
}
