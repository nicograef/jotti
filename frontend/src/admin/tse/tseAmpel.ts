import type { TSESignaturQueue, TSEStatus } from './TSEBackend'

// Deckt sich mit der Nachsigniert-Schwelle im Backend.
const RUECKSTAND_WARN_SEKUNDEN = 60

export const TSE_TEST_WARNUNG =
  'Die TSE läuft in der Test-Umgebung: Belege sind steuerlich nicht gültig. Vor dem Kassieren auf LIVE umstellen.'

// TEST-Signaturen erfüllen § 146a AO nicht; die Umgebung ist im Schema ein freier String.
export function istTestUmgebung(tseStatus: TSEStatus | undefined): boolean {
  return tseStatus?.istKonfiguriert === true && tseStatus.umgebung === 'TEST'
}

export interface TSEAmpel {
  fehler: boolean
  ueberschrift: string
  nichtKonfiguriert: boolean
  rueckstand: boolean
  signaturFehlgeschlagen: boolean
  testUmgebung: boolean
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
  const testUmgebung = !nichtKonfiguriert && istTestUmgebung(tseStatus)
  const fehler =
    nichtKonfiguriert || rueckstand || signaturFehlgeschlagen || testUmgebung

  let ueberschrift: string
  if (nichtKonfiguriert) {
    ueberschrift = 'TSE ist nicht eingerichtet'
  } else if (signaturFehlgeschlagen || rueckstand) {
    ueberschrift = 'TSE braucht Aufmerksamkeit'
  } else if (testUmgebung) {
    ueberschrift = 'TSE läuft in der Test-Umgebung'
  } else {
    ueberschrift = 'Ja — TSE signiert normal'
  }

  return {
    fehler,
    ueberschrift,
    nichtKonfiguriert,
    rueckstand,
    signaturFehlgeschlagen,
    testUmgebung,
  }
}
