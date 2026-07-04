import { TriangleAlert } from 'lucide-react'
import { useState } from 'react'
import { NavLink } from 'react-router'

import { useTSESignaturQueue, useTSEStatus } from '@/admin/settings/hooks'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

import { DsfinvkExportButton } from './DsfinvkExportButton'
import { useKassensitzungen, useLiveReporting, useReport } from './hooks'
import { LiveReportingSection } from './LiveReportingSection'
import { ReportingFilter } from './ReportingFilter'
import { ReportingResults } from './ReportingResults'

// Ab rund einer Minute Rückstand warnt das Dashboard (deckt sich mit der
// Nachsigniert-Schwelle im Backend); der Störungszeitraum entsteht erst ab
// zwei Minuten.
const RUECKSTAND_WARN_SEKUNDEN = 60

export function AdminDashboardPage() {
  const { liveData, isPending: liveLoading } = useLiveReporting()
  const { kassensitzungen, isPending: listLoading } = useKassensitzungen()
  const { tseStatus, isPending: tseLoading } = useTSEStatus()
  const { queue } = useTSESignaturQueue()
  const [selectedNr, setSelectedNr] = useState<number | null>(null)

  const effectiveNr = selectedNr ?? kassensitzungen.at(0)?.zNr ?? null
  const { result, isPending: reportLoading } = useReport(effectiveNr)

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

  return (
    <>
      {showTSEBanner && (
        <Alert variant="destructive" className="mb-6">
          <TriangleAlert className="size-4" />
          <AlertTitle>TSE prüfen</AlertTitle>
          <AlertDescription>
            {showKonfigWarnung && <span>Die TSE ist nicht konfiguriert. </span>}
            {rueckstand && (
              <span>
                {queue?.offeneAuftraege} Vorgänge warten auf Signatur.{' '}
              </span>
            )}
            {fehlgeschlagen && (
              <span>
                {queue?.fehlgeschlageneAuftraege} Vorgänge konnten nicht
                signiert werden
                {queue?.letzterFehler ? ` (${queue.letzterFehler})` : ''}.{' '}
              </span>
            )}
            Mehr dazu unter{' '}
            <NavLink
              to="/admin/finanzamt"
              className="underline underline-offset-4"
            >
              Finanzamt
            </NavLink>
            .
          </AlertDescription>
        </Alert>
      )}

      <LiveReportingSection liveData={liveData} loading={liveLoading} />
      <hr className="my-8" />

      <h2 className="mt-10 text-lg font-semibold">Historische Auswertung</h2>
      <div className="mt-4 flex flex-wrap items-center gap-3">
        <ReportingFilter
          kassensitzungen={kassensitzungen}
          kassensitzungNr={effectiveNr}
          loading={listLoading}
          onKassensitzungNrChange={setSelectedNr}
        />
        <DsfinvkExportButton kassensitzungNr={effectiveNr} />
      </div>

      {result && (
        <div className="my-6">
          <ReportingResults result={result} loading={reportLoading} />
        </div>
      )}
    </>
  )
}
