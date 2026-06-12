import { Button } from '@/components/ui/button'
import { formatCents } from '@/lib/utils'

interface StickyActionBarProps extends React.ComponentProps<'div'> {
  label: string
  anzahl: number
  summeCents: number
  disabled?: boolean
}

/**
 * Floating Primäraktion am unteren Viewport-Rand. Auf Mobilgeräten sitzt sie
 * oberhalb der fixierten Tab-Leiste (TablePage), auf Desktop am unteren Rand –
 * der `md`-Breakpoint deckt sich mit `useIsMobile` (768px).
 *
 * Die Leiste folgt dem bestehenden DrawerTrigger-Muster: Der `DrawerTrigger`
 * klont das Wurzel-`div` (Ref und Click-Handler landen via `...props` darauf);
 * der innere Button spiegelt nur den deaktivierten Zustand. Das Öffnen bei leerer
 * Auswahl wird vom `onOpenChange` des Drawers abgefangen (Guard), nicht hier.
 */
export function StickyActionBar({
  label,
  anzahl,
  summeCents,
  disabled,
  ...props
}: StickyActionBarProps) {
  return (
    <div
      className="fixed inset-x-0 bottom-[calc(4rem+env(safe-area-inset-bottom,0px))] z-40 flex justify-center px-4 md:bottom-[calc(1rem+env(safe-area-inset-bottom,0px))]"
      {...props}
    >
      <Button
        disabled={disabled}
        className="flex h-14 w-full max-w-md items-center justify-between gap-3 text-base shadow-lg"
      >
        <span className="flex items-center gap-2">
          <span className="rounded-full bg-primary-foreground/20 px-2 py-0.5 text-sm font-semibold tabular-nums">
            {anzahl}
          </span>
          {label}
        </span>
        <span className="font-bold tabular-nums">
          {formatCents(summeCents)}&nbsp;€
        </span>
      </Button>
    </div>
  )
}
