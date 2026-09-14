import { useAktiveKassensitzung, useKassenbestand } from '@/admin/kasse/hooks'
import { KassensitzungStatus } from '@/admin/kasse/Kassensitzung'
import { beschreibeFehlBons } from '@/admin/settings/DruckstationBackend'
import { useFehlgeschlageneDruckauftraege } from '@/admin/settings/hooks'
import { useTSESignaturQueue, useTSEStatus } from '@/admin/tse/hooks'
import { tseAmpel } from '@/admin/tse/tseAmpel'
import { formatEuro } from '@/lib/utils'

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
  const { kassensitzung } = useAktiveKassensitzung()
  const { kassenbestand } = useKassenbestand(kassensitzung?.zNr ?? null)
  const { tseStatus, isPending: tseLoading } = useTSEStatus()
  const { queue } = useTSESignaturQueue()
  const { druckauftraege } = useFehlgeschlageneDruckauftraege()

  const {
    fehler: tseFehler,
    nichtKonfiguriert: tseNichtKonfiguriert,
    rueckstand,
    signaturFehlgeschlagen,
  } = tseAmpel(tseStatus, tseLoading, queue)

  const tseText = tseNichtKonfiguriert
    ? 'Nicht konfiguriert'
    : signaturFehlgeschlagen
      ? `${String(queue?.fehlgeschlageneAuftraege ?? 0)} Vorgänge konnten nicht signiert werden`
      : rueckstand
        ? `${String(queue?.offeneAuftraege ?? 0)} Vorgänge warten auf Signatur`
        : `${String(queue?.offeneAuftraege ?? 0)} Vorgänge in Warteschlange (normal)`

  const druckFehler = druckauftraege.length > 0
  const { singular: druckSingular, plural: druckPlural } = beschreibeFehlBons(
    druckauftraege.map((auftrag) => auftrag.bonArt),
  )
  const druckTitel = druckFehler
    ? druckauftraege.length === 1
      ? `1 ${druckSingular} nicht gedruckt`
      : `${String(druckauftraege.length)} ${druckPlural} nicht gedruckt`
    : 'Drucker bereit'
  const druckText = druckFehler ? 'Drucker prüfen' : 'Alle Bons gedruckt'

  // Beide Angaben stammen aus eigenen Queries; fehlt eine, entfällt nur ihr
  // Teil (kein hängendes „seit " ohne Zeit).
  const kasseTeile = [
    kassensitzung &&
      `Kassentag seit ${formatStand(new Date(kassensitzung.eroeffnetAm).getTime())}`,
    kassenbestand !== null &&
      `Soll-Bestand ${formatEuro(kassenbestand.sollBestandCents)}`,
  ].filter((teil): teil is string => typeof teil === 'string')
  const kasseText = kasseTeile.length > 0 ? kasseTeile.join(' · ') : 'geöffnet'

  const abschlussUnterbrochen =
    kassensitzung?.status === KassensitzungStatus.WIRD_ABGESCHLOSSEN
  const kasseTitel = abschlussUnterbrochen
    ? 'Abschluss unterbrochen'
    : 'Kasse offen'

  return (
    <LiveReportingSection
      liveData={liveData}
      loading={liveLoading}
      dataUpdatedAt={dataUpdatedAt}
      onRefresh={() => void refetch()}
      statusZeile={
        liveData !== null && (
          <UebersichtStatusZeile
            kasseTitel={kasseTitel}
            kasseFehler={abschlussUnterbrochen}
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
