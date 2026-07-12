import { cn } from '@/lib/utils'

// Ampel-Zustand eines Statuspunkts: ok = grün (--primary), fehler = rot
// (--destructive), neutral = gedämpft. Reiner Design-Baustein ohne eigene
// Logik; die Zuordnung trifft die aufrufende Stelle.
export type StatusDotZustand = 'ok' | 'fehler' | 'neutral'

const zustandKlasse: Record<StatusDotZustand, string> = {
  ok: 'bg-primary',
  fehler: 'bg-destructive',
  neutral: 'bg-muted-foreground',
}

// 7-px-Punkt gemäß Design-Handoff (Abschnitt 0). label wird als aria-label
// gesetzt, damit der Zustand auch ohne Sichtkontakt lesbar ist. role="img"
// statt role="status": Der Punkt trägt Bedeutung (roter Punkt = Problem), soll
// aber nicht als Live-Region bei jedem Refetch neu vorgelesen werden.
export function StatusDot({
  zustand,
  label,
  className,
}: {
  zustand: StatusDotZustand
  label: string
  className?: string
}) {
  return (
    <span
      role="img"
      aria-label={label}
      className={cn(
        'inline-block size-[7px] shrink-0 rounded-full',
        zustandKlasse[zustand],
        className,
      )}
    />
  )
}
