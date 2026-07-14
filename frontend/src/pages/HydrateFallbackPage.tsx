import { Wortmarke } from '@/components/common/Wortmarke'
import { Spinner } from '@/components/ui/spinner'

// Wird von React Router beim Erstladen gerendert, solange der initiale
// Loader noch läuft. Ohne diese Komponente warnt React Router in der
// Konsole, dass kein HydrateFallback definiert ist.
export function HydrateFallbackPage() {
  return (
    <div className="flex flex-col min-h-screen items-center justify-center gap-4 bg-primary/5">
      <Wortmarke as="h1" className="text-[38px] text-center" />
      <Spinner className="size-6" />
    </div>
  )
}
