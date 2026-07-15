import { formatCents } from '@/lib/utils'

// „Nach dieser Zahlung noch offen"-Zeile. Unter lg rendert der Handy-Container
// sie in den Dock-Slot, ab lg trägt die Abschluss-Spalte sie selbst im Footer —
// dieselbe Darstellung, nur ein anderer Ort.
export function RestbetragZeile({ cents }: { cents: number }) {
  return (
    <div className="flex items-center justify-between gap-3 text-[13px] text-muted-foreground">
      <span>Nach dieser Zahlung noch offen</span>
      <span className="font-semibold tabular-nums text-foreground">
        {formatCents(cents)}&nbsp;€
      </span>
    </div>
  )
}
