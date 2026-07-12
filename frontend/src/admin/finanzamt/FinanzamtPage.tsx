import { AdminPageHeader } from '../components/AdminPageHeader'
import { BetreiberSection } from './BetreiberSection'
import { DokumenteUndPflichtenSection } from './DokumenteUndPflichtenSection'
import { KassenidentitaetSection } from './KassenidentitaetSection'
import { SignaturauftraegeSection } from './SignaturauftraegeSection'
import { TSEAnbindungSection } from './TSEAnbindungSection'
import { TSEAusfalldokumentationSection } from './TSEAusfalldokumentationSection'

export function FinanzamtPage() {
  return (
    <div className="flex flex-col gap-6 max-w-2xl">
      <AdminPageHeader
        titel="Finanzamt & TSE"
        unterzeile="Einmal einrichten, dann läuft es im Hintergrund. jotti erinnert dich, wenn etwas fehlt."
      />
      <BetreiberSection />
      <KassenidentitaetSection />
      <SignaturauftraegeSection />
      <TSEAusfalldokumentationSection />
      <TSEAnbindungSection />
      <DokumenteUndPflichtenSection />
    </div>
  )
}
