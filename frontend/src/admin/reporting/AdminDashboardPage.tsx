import { useKassenbestand, useOffeneKassensitzung } from '@/admin/kasse/hooks'
import { useFehlgeschlageneDruckauftraege } from '@/admin/settings/hooks'
import {
  RUECKSTAND_WARN_SEKUNDEN,
  useTSESignaturQueue,
  useTSEStatus,
} from '@/admin/tse/hooks'
import { formatCents } from '@/lib/utils'

import { useLiveReporting } from './hooks'
import { LiveReportingSection } from './LiveReportingSection'
import { UebersichtStatusZeile } from './UebersichtStatusZeile'
import { formatStand } from './utils'

export function AdminDashboardPage() {
  const {
    liveData,
    isPending: liveLoading,
    dataUpdatedAt,
    refetch,
  } = useLiveReporting()
  const { kassensitzung } = useOffeneKassensitzung()
  const { kassenbestand } = useKassenbestand(kassensitzung?.zNr ?? null)
  const { tseStatus, isPending: tseLoading } = useTSEStatus()
  const { queue } = useTSESignaturQueue()
  const { druckauftraege } = useFehlgeschlageneDruckauftraege()

  // TSE-Fehler und -Schwellen wie im bisherigen Banner: ohne Konfiguration der
  // permanente Alarm, sonst Rückstand über der geteilten Schwelle oder
  // endgültig fehlgeschlagene Signaturen.
  const tseNichtKonfiguriert = !tseLoading && !tseStatus?.istKonfiguriert
  const rueckstand =
    !tseNichtKonfiguriert &&
    (queue?.rueckstandSekunden ?? 0) >= RUECKSTAND_WARN_SEKUNDEN
  const signaturFehlgeschlagen =
    !tseNichtKonfiguriert && (queue?.fehlgeschlageneAuftraege ?? 0) > 0
  const tseFehler = tseNichtKonfiguriert || rueckstand || signaturFehlgeschlagen

  const tseText = tseNichtKonfiguriert
    ? 'Nicht konfiguriert'
    : signaturFehlgeschlagen
      ? `${String(queue?.fehlgeschlageneAuftraege ?? 0)} Vorgänge konnten nicht signiert werden`
      : rueckstand
        ? `${String(queue?.offeneAuftraege ?? 0)} Vorgänge warten auf Signatur`
        : `${String(queue?.offeneAuftraege ?? 0)} Vorgänge in Warteschlange (normal)`

  const druckFehler = druckauftraege.length > 0
  const druckTitel = druckFehler
    ? druckauftraege.length === 1
      ? '1 Bon nicht gedruckt'
      : `${String(druckauftraege.length)} Bons nicht gedruckt`
    : 'Drucker bereit'
  const druckText = druckFehler ? 'Drucker prüfen' : 'Alle Bons gedruckt'

  // „seit HH:MM" aus dem Eröffnungszeitpunkt plus Soll-Bestand der offenen
  // Sitzung. Beide Angaben stammen aus eigenen Queries und können noch fehlen;
  // fehlt die eine, entfällt nur ihr Teil (kein hängendes „seit " ohne Zeit).
  const kasseTeile = [
    kassensitzung &&
      `seit ${formatStand(new Date(kassensitzung.eroeffnetAm).getTime())}`,
    kassenbestand !== null &&
      `Soll-Bestand ${formatCents(kassenbestand.sollBestandCents)} €`,
  ].filter((teil): teil is string => teil !== false && teil !== '')
  const kasseText = kasseTeile.length > 0 ? kasseTeile.join(' · ') : 'geöffnet'

  return (
    <LiveReportingSection
      liveData={liveData}
      loading={liveLoading}
      dataUpdatedAt={dataUpdatedAt}
      onRefresh={() => void refetch()}
      statusZeile={
        liveData !== null && (
          <UebersichtStatusZeile
            kasseText={kasseText}
            tseFehler={tseFehler}
            tseText={tseText}
            druckFehler={druckFehler}
            druckTitel={druckTitel}
            druckText={druckText}
          />
        )
      }
    />
  )
}
