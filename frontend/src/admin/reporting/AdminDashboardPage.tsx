import { TriangleAlert } from 'lucide-react'
import { NavLink } from 'react-router'

import { useFehlgeschlageneDruckauftraege } from '@/admin/settings/hooks'
import { useTSESignaturQueue, useTSEStatus } from '@/admin/tse/hooks'
import { Alert, AlertTitle } from '@/components/ui/alert'

import { useLiveReporting } from './hooks'
import { LiveReportingSection } from './LiveReportingSection'

// Ab rund einer Minute Rückstand warnt das Dashboard (deckt sich mit der
// Nachsigniert-Schwelle im Backend); der Störungszeitraum entsteht erst ab
// zwei Minuten.
const RUECKSTAND_WARN_SEKUNDEN = 60

export function AdminDashboardPage() {
  const {
    liveData,
    isPending: liveLoading,
    dataUpdatedAt,
    refetch,
  } = useLiveReporting()
  const { tseStatus, isPending: tseLoading } = useTSEStatus()
  const { queue } = useTSESignaturQueue()
  const { druckauftraege } = useFehlgeschlageneDruckauftraege()

  // Ohne TSE-Konfiguration steht der permanente Konfigurationsalarm; ein
  // Queue-Alarm (Rückstand oder endgültig fehlgeschlagene Aufträge) erscheint
  // nur, solange die TSE konfiguriert ist.
  const showKonfigWarnung = !tseLoading && !tseStatus?.istKonfiguriert
  const rueckstand =
    !showKonfigWarnung &&
    (queue?.rueckstandSekunden ?? 0) >= RUECKSTAND_WARN_SEKUNDEN
  const fehlgeschlagen =
    !showKonfigWarnung && (queue?.fehlgeschlageneAuftraege ?? 0) > 0
  const showQueueWarnung = rueckstand || fehlgeschlagen
  const showTSEBanner = showKonfigWarnung || showQueueWarnung

  // Drucker-Ausfälle proaktiv melden: fehlgeschlagene Druckaufträge sind sonst
  // nur auf der Druckstationen-Unterseite sichtbar.
  const showDruckBanner = druckauftraege.length > 0

  return (
    <>
      {showDruckBanner && (
        <Alert variant="destructive" className="mb-3">
          <TriangleAlert className="size-4" />
          <AlertTitle className="font-normal">
            {druckauftraege.length === 1
              ? '1 Druckauftrag konnte nicht gedruckt werden.'
              : `${String(druckauftraege.length)} Druckaufträge konnten nicht gedruckt werden.`}{' '}
            <NavLink to="/admin/druckstationen" className="font-medium">
              Druckstationen
            </NavLink>
          </AlertTitle>
        </Alert>
      )}

      {showTSEBanner && (
        <Alert variant="destructive" className="mb-3">
          <TriangleAlert className="size-4" />
          <AlertTitle className="font-normal">
            {showKonfigWarnung && <span>Die TSE ist nicht konfiguriert. </span>}
            {rueckstand && (
              <span>
                {queue?.offeneAuftraege === 1
                  ? '1 Vorgang wartet auf Signatur.'
                  : `${String(queue?.offeneAuftraege ?? 0)} Vorgänge warten auf Signatur.`}{' '}
              </span>
            )}
            {fehlgeschlagen && (
              <span>
                {queue?.fehlgeschlageneAuftraege === 1
                  ? '1 Vorgang konnte nicht signiert werden'
                  : `${String(queue?.fehlgeschlageneAuftraege ?? 0)} Vorgänge konnten nicht signiert werden`}
                {queue?.letzterFehler ? ` (${queue.letzterFehler})` : ''}.{' '}
              </span>
            )}
            <NavLink to="/admin/finanzamt" className="font-medium">
              Finanzamt
            </NavLink>
          </AlertTitle>
        </Alert>
      )}

      <LiveReportingSection
        liveData={liveData}
        loading={liveLoading}
        dataUpdatedAt={dataUpdatedAt}
        onRefresh={() => void refetch()}
      />
    </>
  )
}
