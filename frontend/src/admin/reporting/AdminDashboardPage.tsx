import { useReporting } from './hooks'
import { ReportingFilter } from './ReportingFilter'
import { ReportingResults } from './ReportingResults'

export function AdminDashboardPage() {
  const reporting = useReporting()

  return (
    <>
      <h1 className="text-2xl font-bold">Reporting</h1>
      <div className="mt-4">
        <ReportingFilter
          kassensitzungNr={reporting.kassensitzungNr}
          loading={reporting.loading}
          onKassensitzungNrChange={reporting.setKassensitzungNr}
          onAuswerten={reporting.auswerten}
        />
      </div>

      {reporting.result && (
        <div className="mt-6">
          <ReportingResults result={reporting.result} />
        </div>
      )}
    </>
  )
}
