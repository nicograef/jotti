import { QueryCache, QueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { BackendError, ResponseBodyError } from './Backend'
import { appendReferenz } from './errorMessages'

// Danach ist die Störung dauerhaft genug, um sie der Helferin zu melden.
const MAX_WIEDERHOLUNGEN = 2

const queryFehlerMeldung =
  'Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.'

// Als `meta` einer Query gesetzt, unterdrückt dieses Flag den globalen
// Fehler-Toast: Eine dauerhaft weiterlaufende Hintergrundabfrage würde im
// Funkloch sonst alle 30 Sekunden melden, ohne dass jemand reagieren kann.
export const OHNE_FEHLER_TOAST = { ohneFehlerToast: true }

// Wiederholt nur, was beim nächsten Versuch anders ausgehen kann: Netzfehler
// und Serverfehler ab 500. Ein 4xx (Validierung, Berechtigung, Konflikt) und
// ein ResponseBodyError stehen schon beim ersten Versuch fest.
function sollWiederholen(
  bisherigeWiederholungen: number,
  error: unknown,
): boolean {
  if (bisherigeWiederholungen >= MAX_WIEDERHOLUNGEN) {
    return false
  }

  if (error instanceof BackendError) {
    return error.status >= 500
  }

  if (error instanceof ResponseBodyError) {
    return false
  }

  // Alles Übrige kam ohne lesbare Antwort zurück — ein abgebrochener fetch
  // wirft einen nackten TypeError. Im Vereins-WLAN ist das meist vorübergehend.
  return true
}

// Ohne diesen Fehler-Handler verschwinden Query-Fehler stumm und die Seiten
// zeigen Leer-Defaults (z. B. Saldo 0,00 €). Die feste Toast-ID bündelt
// mehrere gleichzeitig fehlschlagende Queries zu einem Toast.
export function createQueryClient(): QueryClient {
  return new QueryClient({
    // Nur für Lese-Queries: Ein wiederholter Schreibvorgang würde doppelt
    // buchen. Buchungen laufen über useActionSubmit an react-query vorbei;
    // einzige Mutation ist der DSFinV-K-Export.
    defaultOptions: {
      queries: {
        retry: sollWiederholen,
      },
    },
    queryCache: new QueryCache({
      onError: (error, query) => {
        console.error(error)

        if (query.meta?.ohneFehlerToast === true) return

        const referenz =
          error instanceof BackendError ? error.referenz : undefined
        toast.error(appendReferenz(queryFehlerMeldung, referenz), {
          id: 'query-fehler',
        })
      },
    }),
  })
}
