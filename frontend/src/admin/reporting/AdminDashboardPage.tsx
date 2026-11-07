import { useAktiveKassensitzung, useKassenbestand } from '@/admin/kasse/hooks'
import { beschreibeFehlBons } from '@/admin/settings/DruckstationBackend'
import { useFehlgeschlageneDruckauftraege } from '@/admin/settings/hooks'
import { useTSEAmpel } from '@/admin/tse/hooks'
import { TSE_TEST_WARNUNG } from '@/admin/tse/tseAmpel'
import { formatAnzahl,formatEuro, formatUhrzeit } from '@/lib/utils'

import { useLiveReporting } from './hooks'
import { LiveReportingSection } from './LiveReportingSection'
import { UebersichtStatusZeile } from './UebersichtStatusZeile'

export function AdminDashboardPage() {
  const {
    liveData,
    isPending: liveLoading,
    isError: liveFehler,
    dataUpdatedAt,
    refetch,
  } = useLiveReporting()
  const { kassensitzung, abschlussUnterbrochen } = useAktiveKassensitzung()
  const { kassenbestand } = useKassenbestand(kassensitzung?.zNr ?? null)
  const { druckauftraege } = useFehlgeschlageneDruckauftraege()

  const {
    fehler: tseFehler,
    nichtKonfiguriert: tseNichtKonfiguriert,
    rueckstand,
    signaturFehlgeschlagen,
    testUmgebung,
    queue,
  } = useTSEAmpel()

  const tseText = tseNichtKonfiguriert
    ? 'Nicht konfiguriert'
    : signaturFehlgeschlagen
      ? `${formatAnzahl(queue?.fehlgeschlageneAuftraege ?? 0, 'Vorgang konnte', 'Vorgänge konnten')} nicht signiert werden`
      : rueckstand
        ? `${formatAnzahl(queue?.offeneAuftraege ?? 0, 'Vorgang wartet', 'Vorgänge warten')} auf Signatur`
        : testUmgebung
          ? TSE_TEST_WARNUNG
          : `${formatAnzahl(queue?.offeneAuftraege ?? 0, 'Vorgang', 'Vorgänge')} in Warteschlange (normal)`

  const druckFehler = druckauftraege.length > 0
  const { singular: druckSingular, plural: druckPlural } = beschreibeFehlBons(
    druckauftraege.map((auftrag) => auftrag.bonArt),
  )
  const druckTitel = druckFehler
    ? `${formatAnzahl(druckauftraege.length, druckSingular, druckPlural)} nicht gedruckt`
    : 'Drucker bereit'
  const druckText = druckFehler ? 'Drucker prüfen' : 'Alle Bons gedruckt'

  // Beide Angaben stammen aus eigenen Queries; fehlt eine, entfällt nur ihr
  // Teil (kein hängendes „seit " ohne Zeit).
  const kasseTeile = [
    kassensitzung &&
      `Kassentag seit ${formatUhrzeit(kassensitzung.eroeffnetAm)}`,
    kassenbestand !== null &&
      `Soll-Bestand ${formatEuro(kassenbestand.sollBestandCents)}`,
  ].filter((teil): teil is string => typeof teil === 'string')
  const kasseText = kasseTeile.length > 0 ? kasseTeile.join(' · ') : 'geöffnet'

  const kasseTitel = abschlussUnterbrochen
    ? 'Abschluss unterbrochen'
    : 'Kasse offen'

  return (
    <LiveReportingSection
      liveData={liveData}
      loading={liveLoading}
      fehler={liveFehler}
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
