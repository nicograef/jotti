import { BetreiberSection } from '@/admin/finanzamt/BetreiberSection'
import { KassenidentitaetSection } from '@/admin/finanzamt/KassenidentitaetSection'
import { TSEAusfalldokumentationSection } from '@/admin/finanzamt/TSEAusfalldokumentationSection'
import { TSEKonfigurationSection } from '@/admin/tse/TSEKonfigurationSection'

export function EinstellungenPage() {
  return (
    <div className="flex flex-col gap-6 max-w-2xl">
      <BetreiberSection />
      <KassenidentitaetSection />
      <TSEKonfigurationSection />
      <TSEAusfalldokumentationSection />
    </div>
  )
}
