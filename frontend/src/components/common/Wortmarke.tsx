import type { ElementType } from 'react'

import { cn } from '@/lib/utils'

interface WortmarkeProps {
  // HTML-Element der Einsatzstelle. Erhält die bestehende Semantik: wo heute
  // ein h1 steht, wird `as="h1"` übergeben; der mobile Admin-Kopf bleibt span.
  as?: ElementType
  className?: string
}

// Die Wortmarke „jotti" als selektierbarer Text (kein Bild) in Space Grotesk
// 700 mit dem Spektralverlauf als Text-Füllung. Größe und Ausrichtung kommen
// über className der Einsatzstelle; der Verlauf folgt automatisch dem Theme,
// weil --spectral im Dark Mode überschrieben ist.
export function Wortmarke({
  as: Component = 'span',
  className,
}: WortmarkeProps) {
  return (
    <Component
      className={cn(
        'bg-[image:var(--spectral)] bg-clip-text font-heading font-bold text-transparent',
        className,
      )}
    >
      jotti
    </Component>
  )
}
