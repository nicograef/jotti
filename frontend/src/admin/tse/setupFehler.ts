import { toast } from 'sonner'

import type { BackendError } from '@/lib/Backend'

import {
  type TSEEinrichtenErgebnis,
  unvollstaendigeEinrichtung,
} from './TSEBackend'

// Fehlermeldungen der fiskaly-Einrichtung, geteilt von den Wizard-Schritten.
export const ZUGANGSDATEN_FEHLER = {
  tse_setup_zugangsdaten_ungueltig:
    'API-Key oder API-Secret ist ungültig. Bitte Zugangsdaten prüfen.',
}
export const SETUP_FEHLER = {
  ...ZUGANGSDATEN_FEHLER,
  tse_setup_umgebung_abweichung:
    'Die Umgebung der Zugangsdaten hat sich geändert. Bitte das Konto erneut prüfen.',
  tse_konfiguration_kassensitzung_offen:
    'Die TSE kann nicht geändert werden, solange eine Kassensitzung aktiv ist (offen oder Abschluss unterbrochen). Bitte zuerst den Kassenabschluss durchführen bzw. wiederholen.',
}

// Behandelt tse_einrichtung_unvollstaendig: PUK und PIN dieser Antwort gibt es
// nirgends sonst, ein flüchtiger Toast verlöre sie.
export function unvollstaendigAnzeigen(
  onUnvollstaendig: (ergebnis: TSEEinrichtenErgebnis) => void,
) {
  return {
    tse_einrichtung_unvollstaendig: (error: BackendError) => {
      const ergebnis = unvollstaendigeEinrichtung(error)
      if (ergebnis) {
        onUnvollstaendig(ergebnis)
      } else {
        toast.error(
          'Die Einrichtung ist nicht abgeschlossen. Bitte die TSE erneut prüfen.',
        )
      }
    },
  }
}
