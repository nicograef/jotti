import { Check } from 'lucide-react'
import { useEffect } from 'react'

// Anzeigedauer bis zum automatischen Schließen (Motion-Inventar „Erfolgs-Pop",
// ~1,4 s). Ein Tap schließt jederzeit früher.
const ANZEIGE_DAUER_MS = 1400

interface ErfolgsPopProps {
  open: boolean
  text: string
  onDismiss: () => void
}

// Vollbild-Overlay des sichtbaren Pops: solide Abdunklung als Fallback, darüber
// die feinere color-mix-Tönung für moderne Browser (siehe style-Prop), Blur nur
// dort, wo backdrop-filter unterstützt wird (analog zu dialog/drawer/sheet). Das
// Overlay blockiert bewusst Interaktion, bis es sich schließt: Der nachgelagerte
// Refetch (onDismiss in TablePage) soll greifen, bevor der nächste Tap ein noch
// nicht aktualisiertes Bedienelement trifft.
const OVERLAY_CLASSES =
  'fixed inset-0 z-50 flex flex-col items-center justify-center gap-5 bg-[rgb(0_0_0/0.25)] supports-backdrop-filter:backdrop-blur-[6px]'

// Unübersehbare Buchungsbestätigung im Service: Vollbild-Overlay mit geblurtem
// Backdrop und Häkchen-Kreis in Primärgrün, das nach kurzer Zeit automatisch
// wieder verschwindet. Wird in den drei Buchungsflows (Bestellen, Kassieren,
// Direktverkauf) statt eines Toasts genutzt. Die role="status"-Live-Region ist
// dauerhaft gemountet und wechselt nur ihren Textinhalt, damit Screenreader den
// Erfolg zuverlässig ankündigen (eine frisch befüllt gemountete Region wird oft
// verschluckt). Der Pop meldet das Schließen über onDismiss (Auto-Dismiss oder
// Tap); der nachgelagerte Statuswechsel/Refetch folgt erst dann, sodass sichtbare
// Änderungen dem Pop folgen.
export function ErfolgsPop({ open, text, onDismiss }: ErfolgsPopProps) {
  useEffect(() => {
    if (!open) return
    const timer = setTimeout(onDismiss, ANZEIGE_DAUER_MS)
    return () => {
      clearTimeout(timer)
    }
  }, [open, onDismiss])

  return (
    <div
      role="status"
      className={open ? OVERLAY_CLASSES : 'sr-only'}
      onClick={open ? onDismiss : undefined}
      // Die color-mix-Tönung liegt inline über der soliden bg-Fallback-Klasse:
      // moderne Browser nutzen sie, iOS Safari 16.0–16.1 verwirft sie und
      // behält die sichtbare Abdunklung.
      style={
        open
          ? {
              backgroundColor:
                'color-mix(in oklab, var(--background) 55%, transparent)',
            }
          : undefined
      }
    >
      {open && (
        <>
          {/* Spring-Pop (450 ms) statt der kanonischen 350 ms von animate-pop. */}
          <div className="flex size-[76px] items-center justify-center rounded-full bg-primary text-white ring-8 ring-primary/15 animate-pop [animation-duration:450ms] [animation-timing-function:cubic-bezier(0.34,1.56,0.64,1)]">
            <Check className="size-[34px]" strokeWidth={3} aria-hidden />
          </div>
          <p className="animate-fade-up text-[17px] font-semibold [animation-delay:100ms] [animation-duration:350ms]">
            {text}
          </p>
        </>
      )}
    </div>
  )
}
