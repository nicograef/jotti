import { AdminPageHeader } from '../components/AdminPageHeader'
import { EinrichtungSection } from './EinrichtungSection'
import { GutZuWissenSection } from './GutZuWissenSection'
import { LaeuftAllesSection } from './LaeuftAllesSection'

export function FinanzamtPage() {
  return (
    <div className="flex max-w-4xl flex-col gap-6">
      <AdminPageHeader
        titel="Finanzamt & TSE"
        unterzeile="Einmal einrichten, dann läuft es im Hintergrund. jotti erinnert dich, wenn etwas fehlt."
      />
      <EinrichtungSection />
      <LaeuftAllesSection />
      <GutZuWissenSection />
    </div>
  )
}
