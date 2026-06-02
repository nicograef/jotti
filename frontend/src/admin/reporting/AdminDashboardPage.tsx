import { useState } from 'react'

import { useKassensitzungen, useReport } from './hooks'
import { ReportingFilter } from './ReportingFilter'
import { ReportingResults } from './ReportingResults'

export function AdminDashboardPage() {
  const { kassensitzungen, loading: listLoading } = useKassensitzungen()
  const [selectedNr, setSelectedNr] = useState<number | null>(null)

  const effectiveNr = selectedNr ?? kassensitzungen.at(0)?.zNr ?? null
  const { result, loading: reportLoading } = useReport(effectiveNr)

  return (
    <>
      <h1 className="text-2xl font-bold">Reporting</h1>
      <div className="mt-4">
        <ReportingFilter
          kassensitzungen={kassensitzungen}
          kassensitzungNr={effectiveNr}
          loading={listLoading || reportLoading}
          onKassensitzungNrChange={setSelectedNr}
        />
      </div>

      {result && (
        <div className="mt-6">
          <ReportingResults result={result} />
        </div>
      )}
    </>
  )
}
