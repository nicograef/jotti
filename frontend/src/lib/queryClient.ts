import { QueryCache, QueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { BackendError, ResponseBodyError } from './Backend'
import { appendReferenz } from './errorMessages'

// Höchstens zwei Wiederholungen: Danach ist die Störung dauerhaft genug, um sie
// der Helferin zu melden, statt sie weiter warten zu lassen. Der Abstand
// zwischen den Versuchen wächst (exponentieller Standard-Backoff von
// react-query).
const MAX_WIEDERHOLUNGEN = 2

const queryFehlerMeldung =
  'Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.'

// Als `meta` einer Query gesetzt, unterdrückt dieses Flag den globalen
// Fehler-Toast. Gedacht für Hintergrundabfragen, die dauerhaft weiterlaufen
// (Versionsabfrage): Im Funkloch würden sie sonst alle 30 Sekunden eine rote
// Meldung werfen, ohne dass die Helferin etwas tun könnte.
export const OHNE_FEHLER_TOAST = { ohneFehlerToast: true }

// sollWiederholen wiederholt nur, was beim nächsten Versuch anders ausgehen
// kann: Netzfehler und Serverfehler ab Status 500. Ein BackendError mit 4xx
// (Validierung, fehlende Berechtigung, Konflikt) und ein ResponseBodyError
// (Antwort verletzt das Schema) stehen schon beim ersten Versuch fest und
// würden die Meldung nur verzögern.
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

// createQueryClient baut den globalen QueryClient mit zentralem Fehler-Handling:
// Ohne diesen Handler verschwinden Query-Fehler stumm und die Seiten zeigen
// Leer-Defaults (z. B. Saldo 0,00 €). Die feste Toast-ID sorgt dafür, dass
// mehrere gleichzeitig fehlschlagende Queries (z. B. bei Netzabbruch) nur
// einen Toast erzeugen.
export function createQueryClient(): QueryClient {
  return new QueryClient({
    // Die Wiederholungen gelten ausdrücklich nur für Lese-Queries. Für
    // Mutations wird bewusst nichts gesetzt: Ein wiederholter Schreibvorgang
    // würde doppelt buchen. Buchungen laufen ohnehin über useActionSubmit an
    // react-query vorbei; einzige Mutation ist der DSFinV-K-Export.
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
