import { Check } from 'lucide-react'
import { useEffect } from 'react'

// Anzeigedauer bis zum automatischen Schließen (Motion-Inventar „Erfolgs-Pop",
// ~1,4 s). Ein Tippen schließt jederzeit früher.
const ANZEIGE_DAUER_MS = 1400

interface ErfolgsPopProps {
  open: boolean
  text: string
  onDismiss: () => void
}

// Unübersehbare Buchungsbestätigung im Service: Vollbild-Overlay mit geblurtem
// Backdrop und Häkchen-Kreis in Primärgrün, das nach kurzer Zeit automatisch
// wieder verschwindet. Ersetzt in den drei Buchungsflows (Bestellen, Kassieren,
// Direktverkauf) den bisherigen Erfolgs-Toast. role="status" lässt Screenreader
// den Text wie zuvor den Toast ankündigen. Der Pop meldet das Schließen (Auto
// oder Tap) über onDismiss; der nachgelagerte Statuswechsel/Refetch folgt erst
// dann, sodass sichtbare Änderungen dem Pop folgen.
export function ErfolgsPop({ open, text, onDismiss }: ErfolgsPopProps) {
  useEffect(() => {
    if (!open) return
    const timer = setTimeout(onDismiss, ANZEIGE_DAUER_MS)
    return () => {
      clearTimeout(timer)
    }
  }, [open, onDismiss])

  if (!open) return null

  return (
    <div
      role="status"
      onClick={onDismiss}
      className="fixed inset-0 z-50 flex flex-col items-center justify-center gap-5 bg-[color-mix(in_oklab,var(--background)_55%,transparent)] backdrop-blur-[6px]"
    >
      {/* Spring-Pop (450 ms) statt der kanonischen 350 ms von animate-pop. */}
      <div className="flex size-[76px] items-center justify-center rounded-full bg-primary text-white ring-8 ring-primary/15 animate-pop [animation-duration:450ms] [animation-timing-function:cubic-bezier(0.34,1.56,0.64,1)]">
        <Check className="size-[34px]" strokeWidth={3} aria-hidden />
      </div>
      <p className="animate-fade-up text-[17px] font-semibold [animation-delay:100ms] [animation-duration:350ms]">
        {text}
      </p>
    </div>
  )
}
