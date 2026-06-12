import { TriangleAlert } from 'lucide-react'
import { useState } from 'react'
import { NavLink } from 'react-router'

import { useTSEStatus } from '@/admin/settings/hooks'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

import { useKassensitzungen, useLiveReporting, useReport } from './hooks'
import { LiveReportingSection } from './LiveReportingSection'
import { ReportingFilter } from './ReportingFilter'
import { ReportingResults } from './ReportingResults'

export function AdminDashboardPage() {
  const { liveData, isPending: liveLoading } = useLiveReporting()
  const { kassensitzungen, isPending: listLoading } = useKassensitzungen()
  const { tseStatus, isPending: tseLoading } = useTSEStatus()
  const [selectedNr, setSelectedNr] = useState<number | null>(null)

  const effectiveNr = selectedNr ?? kassensitzungen.at(0)?.zNr ?? null
  const { result, isPending: reportLoading } = useReport(effectiveNr)
  const showTSEWarning = !tseLoading && !tseStatus?.istKonfiguriert
  const offeneNachsignierungen = tseStatus?.offeneNachsignierungen ?? 0
  const showNachsignierWarning = !tseLoading && offeneNachsignierungen > 0

  return (
    <>
      {showTSEWarning && (
        <Alert variant="destructive" className="mb-6">
          <TriangleAlert className="size-4" />
          <AlertTitle>TSE ist nicht konfiguriert</AlertTitle>
          <AlertDescription>
            Ohne TSE darf jotti nur für Test- oder Demozwecke verwendet werden.
            Für den regulären Betrieb musst du die TSE-Zugangsdaten unter{' '}
            <NavLink
              to="/admin/einstellungen"
              className="underline underline-offset-4"
            >
              Einstellungen
            </NavLink>{' '}
            hinterlegen.
          </AlertDescription>
        </Alert>
      )}

      {showNachsignierWarning && (
        <Alert variant="destructive" className="mb-6">
          <TriangleAlert className="size-4" />
          <AlertTitle>TSE-Nachsignierungen offen</AlertTitle>
          <AlertDescription>
            Aktuell warten {offeneNachsignierungen} Vorgänge auf Nachsignierung.
            Bitte TSE-Anbindung prüfen und die Warteschlange beobachten.
            {tseStatus?.umgebung
              ? ` Verbundene Umgebung: ${tseStatus.umgebung}.`
              : ''}
          </AlertDescription>
        </Alert>
      )}

      <LiveReportingSection liveData={liveData} loading={liveLoading} />
      <hr className="my-8" />

      <h2 className="mt-10 text-lg font-semibold">Historische Auswertung</h2>
      <div className="mt-4">
        <ReportingFilter
          kassensitzungen={kassensitzungen}
          kassensitzungNr={effectiveNr}
          loading={listLoading}
          onKassensitzungNrChange={setSelectedNr}
        />
      </div>

      {result && (
        <div className="my-6">
          <ReportingResults result={result} loading={reportLoading} />
        </div>
      )}
    </>
  )
}
