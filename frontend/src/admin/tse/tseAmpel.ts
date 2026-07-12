import { RUECKSTAND_WARN_SEKUNDEN } from './hooks'
import type { TSESignaturQueue, TSEStatus } from './TSEBackend'

// Klartext-Ampel für den TSE-Signaturzustand, wie sie die „Läuft alles?"-Karte
// verwendet: ohne Konfiguration der permanente Alarm, sonst ein Rückstand über
// der geteilten Schwelle (RUECKSTAND_WARN_SEKUNDEN) oder endgültig
// fehlgeschlagene Signaturen. Reine Funktion — keine Hooks, isoliert testbar.
export interface TSEAmpel {
  fehler: boolean
  ueberschrift: string
}

export function tseAmpel(
  tseStatus: TSEStatus | undefined,
  tseLoading: boolean,
  queue: TSESignaturQueue | undefined,
): TSEAmpel {
  const nichtKonfiguriert = !tseLoading && !tseStatus?.istKonfiguriert
  const rueckstand =
    !nichtKonfiguriert &&
    (queue?.rueckstandSekunden ?? 0) >= RUECKSTAND_WARN_SEKUNDEN
  const signaturFehlgeschlagen =
    !nichtKonfiguriert && (queue?.fehlgeschlageneAuftraege ?? 0) > 0
  const fehler = nichtKonfiguriert || rueckstand || signaturFehlgeschlagen

  let ueberschrift: string
  if (nichtKonfiguriert) {
    ueberschrift = 'TSE ist nicht eingerichtet'
  } else if (signaturFehlgeschlagen || rueckstand) {
    ueberschrift = 'TSE braucht Aufmerksamkeit'
  } else {
    ueberschrift = 'Ja — TSE signiert normal'
  }

  return { fehler, ueberschrift }
}
