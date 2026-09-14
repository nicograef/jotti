import { Button } from '@/components/ui/button'
import { formatEuro } from '@/lib/utils'

import { DockActionSlot } from '../ServiceDock'

interface DockActionButtonProps extends React.ComponentProps<'button'> {
  label: string
  anzahl: number
  summeCents: number
  disabled?: boolean
}

/**
 * `DrawerTrigger asChild` legt Ref und Click-Handler via `...props` auf diesen
 * Button; Radix-Context — und damit der Trigger — bleibt über das Portal in den
 * Dock-Slot hinweg erhalten. Das Öffnen bei leerer Auswahl fängt das
 * `onOpenChange` des Drawers ab (Guard), nicht dieser Button.
 */
export function DockActionButton({
  label,
  anzahl,
  summeCents,
  disabled,
  ...props
}: DockActionButtonProps) {
  return (
    <DockActionSlot>
      <Button
        disabled={disabled}
        className="flex h-14 w-full items-center justify-between gap-3 text-base shadow-lg"
        {...props}
      >
        <span className="flex items-center gap-2">
          {/* Der key-Wechsel remountet den Span und startet die pop-Animation
              neu. */}
          <span
            key={anzahl}
            className="animate-pop rounded-full bg-primary-foreground/20 px-2 py-0.5 text-sm font-semibold tabular-nums [animation-duration:250ms]"
          >
            {anzahl}
          </span>
          {label}
        </span>
        <span className="font-bold tabular-nums">{formatEuro(summeCents)}</span>
      </Button>
    </DockActionSlot>
  )
}
