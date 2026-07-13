import { Minus, Plus } from 'lucide-react'

import { Button } from '@/components/ui/button'

interface StepperProps {
  menge: number
  onAdd: () => void
  onRemove: () => void
  addLabel: string
  removeLabel: string
  addDisabled?: boolean
}

// Stepper ist der einheitliche 44-px-Mengen-Wähler des Service-Bereichs. Plus
// ist dauerhaft primär, Minus outline; bei Menge 0 ist Minus regulär deaktiviert
// (abgeblendet und nicht antippbar wie jeder deaktivierte Outline-Button), damit
// der deaktivierte Zustand eindeutig ist. Die Menge in der Mitte
// hat feste Breite, damit der Zustandswechsel keinen Layout-Shift auslöst.
// addDisabled deckelt das Plus dort, wo eine Höchstmenge gilt (z. B. die
// unbezahlte Menge einer Position).
export function Stepper({
  menge,
  onAdd,
  onRemove,
  addLabel,
  removeLabel,
  addDisabled = false,
}: StepperProps) {
  const leer = menge === 0

  return (
    <div className="flex items-center gap-2">
      <Button
        size="icon"
        variant="outline"
        className="size-11 rounded-full"
        aria-label={removeLabel}
        disabled={leer}
        onClick={(e) => {
          e.stopPropagation()
          onRemove()
        }}
      >
        <Minus />
      </Button>
      <span className="w-7 text-center text-[17px] font-bold tabular-nums">
        {menge}
      </span>
      <Button
        size="icon"
        className="size-11 rounded-full"
        aria-label={addLabel}
        disabled={addDisabled}
        onClick={(e) => {
          e.stopPropagation()
          onAdd()
        }}
      >
        <Plus />
      </Button>
    </div>
  )
}
