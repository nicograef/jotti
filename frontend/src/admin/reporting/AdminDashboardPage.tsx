import { Separator } from '@/components/ui/separator'

import { DashboardKpiCards } from './DashboardKpiCards'
import { useDashboard, useTagesabrechnung } from './hooks'
import { TagesabrechnungFilter } from './TagesabrechnungFilter'
import { TagesabrechnungResults } from './TagesabrechnungResults'

export function AdminDashboardPage() {
  const dashboard = useDashboard()
  const tagesabrechnung = useTagesabrechnung()

  return (
    <>
      <DashboardKpiCards data={dashboard.data} loading={dashboard.loading} />

      <Separator className="mt-8" />

      <h2 className="mt-6 text-lg font-semibold">Tagesabrechnung</h2>
      <div className="mt-4">
        <TagesabrechnungFilter
          vonDate={tagesabrechnung.vonDate}
          vonTime={tagesabrechnung.vonTime}
          vonOpen={tagesabrechnung.vonOpen}
          bisDate={tagesabrechnung.bisDate}
          bisTime={tagesabrechnung.bisTime}
          bisOpen={tagesabrechnung.bisOpen}
          loading={tagesabrechnung.loading}
          onVonDateChange={tagesabrechnung.setVonDate}
          onVonTimeChange={tagesabrechnung.setVonTime}
          onVonOpenChange={tagesabrechnung.setVonOpen}
          onBisDateChange={tagesabrechnung.setBisDate}
          onBisTimeChange={tagesabrechnung.setBisTime}
          onBisOpenChange={tagesabrechnung.setBisOpen}
          onAuswerten={tagesabrechnung.auswerten}
        />
      </div>

      {tagesabrechnung.result && (
        <div className="mt-6">
          <TagesabrechnungResults result={tagesabrechnung.result} />
        </div>
      )}
    </>
  )
}
