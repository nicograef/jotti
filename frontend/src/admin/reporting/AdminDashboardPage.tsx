import { useState } from 'react'

import { useKassensitzungen, useLiveReporting, useReport } from './hooks'
import { LiveReportingSection } from './LiveReportingSection'
import { ReportingFilter } from './ReportingFilter'
import { ReportingResults } from './ReportingResults'

export function AdminDashboardPage() {
  const { liveData, loading: liveLoading } = useLiveReporting()
  const { kassensitzungen, loading: listLoading } = useKassensitzungen()
  const [selectedNr, setSelectedNr] = useState<number | null>(null)

  const effectiveNr = selectedNr ?? kassensitzungen.at(0)?.zNr ?? null
  const { result, loading: reportLoading } = useReport(effectiveNr)

  return (
    <>
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
