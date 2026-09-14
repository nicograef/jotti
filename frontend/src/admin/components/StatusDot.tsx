import { cn } from '@/lib/utils'

export type StatusDotZustand = 'ok' | 'fehler' | 'neutral'

const zustandKlasse: Record<StatusDotZustand, string> = {
  ok: 'bg-primary',
  fehler: 'bg-destructive',
  neutral: 'bg-muted-foreground',
}

// role="img" statt role="status": Der Punkt trägt Bedeutung (roter Punkt =
// Problem), soll aber nicht als Live-Region bei jedem Refetch vorgelesen werden.
export function StatusDot({
  zustand,
  label,
  puls = false,
  className,
}: {
  zustand: StatusDotZustand
  label: string
  puls?: boolean
  className?: string
}) {
  return (
    <span
      role="img"
      aria-label={label}
      className={cn(
        'inline-block size-[7px] shrink-0 rounded-full',
        zustandKlasse[zustand],
        puls && 'animate-pulsedot',
        className,
      )}
    />
  )
}
