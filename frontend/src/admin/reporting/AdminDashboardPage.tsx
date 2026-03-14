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
          vonDate={reporting.vonDate}
          vonTime={reporting.vonTime}
          vonOpen={reporting.vonOpen}
          bisDate={reporting.bisDate}
          bisTime={reporting.bisTime}
          bisOpen={reporting.bisOpen}
          loading={reporting.loading}
          onVonDateChange={reporting.setVonDate}
          onVonTimeChange={reporting.setVonTime}
          onVonOpenChange={reporting.setVonOpen}
          onBisDateChange={reporting.setBisDate}
          onBisTimeChange={reporting.setBisTime}
          onBisOpenChange={reporting.setBisOpen}
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
