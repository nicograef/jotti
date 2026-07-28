import { onlineManager, QueryCache, QueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { BackendError, NetzwerkFehler } from './Backend'
import { appendReferenz } from './errorMessages'

// Höchstens zwei Wiederholungen: Danach ist die Störung dauerhaft genug, um sie
// dem Helfer zu melden, statt ihn weiter warten zu lassen. Der Abstand zwischen
// den Versuchen wächst (exponentieller Standard-Backoff von react-query).
const MAX_WIEDERHOLUNGEN = 2

// Aktualitätsschwelle für Stammdaten — Daten, die sich während einer
// Veranstaltung nicht ändern (Produkte, Benutzer, Betreiber, Kassenidentität,
// Druckstationen, TSE-Konfiguration). Nur solche Hooks setzen sie: Sie sparen
// den größten Teil des Traffics im Vereins-WLAN (jedes Entsperren eines Handys
// lädt sonst sämtliche montierten Queries neu), ohne dass jemand mit einem
// veralteten Wert arbeitet. Alles, was Kassen- oder Tischzustand liefert, bleibt
// ohne Schwelle und lädt beim Mount frisch — das Kriterium schlägt die
// Kategorie: Keine Tisch-Query ist ein Stammdatum, weil jede von ihnen
// `saldoCents` der laufenden Kassensitzung mitführt (`useAllTische` in
// admin/tables/hooks.ts, `useAktiveTische` in service/table/hooks.ts).
export const STAMMDATEN_AKTUALITAET_MS = 30_000

const queryFehlerMeldung =
  'Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.'

// Wiederholt wird nur, was Aussicht auf Erfolg hat: Verbindungsfehler und
// Serverfehler ab Status 500. Ein BackendError mit 4xx (Validierung, fehlende
// Berechtigung, Konflikt) und ein ResponseBodyError gelingen beim zweiten
// Versuch nicht plötzlich und verzögern nur die Fehlermeldung.
function sollWiederholen(anzahlFehlversuche: number, error: unknown): boolean {
  if (anzahlFehlversuche >= MAX_WIEDERHOLUNGEN) {
    return false
  }

  if (error instanceof NetzwerkFehler) {
    return true
  }

  return error instanceof BackendError && error.status >= 500
}

// createQueryClient baut den globalen QueryClient mit zentralem Fehler-Handling:
// Ohne diesen Handler verschwinden Query-Fehler stumm und die Seiten zeigen
// Leer-Defaults (z. B. Saldo 0,00 €). Die feste Toast-ID sorgt dafür, dass
// mehrere gleichzeitig fehlschlagende Queries (z. B. bei Netzabbruch) nur
// einen Toast erzeugen.
export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: sollWiederholen,
      },
    },
    queryCache: new QueryCache({
      onError: (error) => {
        console.error(error)

        // Im Funkloch steht die Ursache bereits im Offline-Banner; ein
        // zusätzlicher Toast über einen Ladefehler wäre hier irreführend.
        if (!onlineManager.isOnline()) {
          return
        }

        const referenz =
          error instanceof BackendError ? error.referenz : undefined
        toast.error(appendReferenz(queryFehlerMeldung, referenz), {
          id: 'query-fehler',
        })
      },
    }),
  })
}
