import { Spinner } from '@/components/ui/spinner'

// Wird von React Router beim Erstladen gerendert, solange der initiale
// Loader noch läuft. Ohne diese Komponente warnt React Router in der
// Konsole, dass kein HydrateFallback definiert ist.
export function HydrateFallbackPage() {
  return (
    <div className="flex flex-col min-h-screen items-center justify-center gap-4 bg-primary/5">
      <h1 className="text-4xl text-center font-extrabold">jotti</h1>
      <Spinner className="size-6" />
    </div>
  )
}
