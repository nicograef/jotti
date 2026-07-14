import { Button } from '@/components/ui/button'
import { formatCents } from '@/lib/utils'

import { DockActionSlot } from '../ServiceDock'

interface DockActionButtonProps extends React.ComponentProps<'button'> {
  label: string
  anzahl: number
  summeCents: number
  disabled?: boolean
}

/**
 * Primäraktion des Service-Docks. Sie rendert per Portal in den Aktions-Slot
 * von ServiceDock (DockActionSlot); die Positionierung liegt beim Dock, nicht
 * hier.
 *
 * Der Button folgt dem bestehenden DrawerTrigger-Muster: `DrawerTrigger asChild`
 * legt Ref und Click-Handler via `...props` auf diesen Button (React 19 reicht
 * `ref` als reguläre Prop durch). Radix-Context — und damit der Trigger — bleibt
 * über das Portal hinweg erhalten. Das Öffnen bei leerer Auswahl fängt das
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
          {/* Die Mengen-Pill poppt bei jeder Mengenänderung: der key-Wechsel
              remountet den Span, wodurch die pop-Animation neu startet. 250 ms
              statt der kanonischen 350 ms gemäß Handoff-Delta. */}
          <span
            key={anzahl}
            className="animate-pop rounded-full bg-primary-foreground/20 px-2 py-0.5 text-sm font-semibold tabular-nums [animation-duration:250ms]"
          >
            {anzahl}
          </span>
          {label}
        </span>
        <span className="font-bold tabular-nums">
          {formatCents(summeCents)}&nbsp;€
        </span>
      </Button>
    </DockActionSlot>
  )
}
