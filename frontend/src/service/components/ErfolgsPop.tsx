import { Check } from 'lucide-react'
import { useEffect } from 'react'

// Anzeigedauer bis zum automatischen Schließen (Motion-Inventar „Erfolgs-Pop",
// ~1,4 s). Der Auto-Dismiss-Timer ist die alleinige Schließ-Logik; ein Tap wird
// bewusst nicht mehr abgefangen (pointer-events-none), damit er sein Ziel unter
// dem Overlay erreicht.
const ANZEIGE_DAUER_MS = 1400

interface ErfolgsPopProps {
  open: boolean
  text: string
  onDismiss: () => void
}

// Vollbild-Overlay des sichtbaren Pops: solide Abdunklung als Fallback, darüber
// die feinere color-mix-Tönung für moderne Browser (siehe style-Prop), Blur nur
// dort, wo backdrop-filter unterstützt wird (analog zu dialog/drawer/sheet).
// pointer-events-none lässt Taps zum darunterliegenden Bedienelement durch.
const OVERLAY_CLASSES =
  'pointer-events-none fixed inset-0 z-50 flex flex-col items-center justify-center gap-5 bg-[rgb(0_0_0/0.25)] supports-backdrop-filter:backdrop-blur-[6px]'

// Unübersehbare Buchungsbestätigung im Service: Vollbild-Overlay mit geblurtem
// Backdrop und Häkchen-Kreis in Primärgrün, das nach kurzer Zeit automatisch
// wieder verschwindet. Wird in den drei Buchungsflows (Bestellen, Kassieren,
// Direktverkauf) statt eines Toasts genutzt. Die role="status"-Live-Region ist
// dauerhaft gemountet und wechselt nur ihren Textinhalt, damit Screenreader den
// Erfolg zuverlässig ankündigen (eine frisch befüllt gemountete Region wird oft
// verschluckt). Der Pop meldet das Schließen über onDismiss (Auto-Dismiss); der
// nachgelagerte Statuswechsel/Refetch folgt erst dann, sodass sichtbare
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
