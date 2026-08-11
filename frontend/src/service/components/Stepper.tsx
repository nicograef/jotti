import { Minus, Plus } from 'lucide-react'

import { Button } from '@/components/ui/button'

interface StepperProps {
  menge: number
  onAdd: () => void
  onRemove: () => void
  addLabel: string
  removeLabel: string
  addDisabled?: boolean
  minusNurAbEins?: boolean
}

// Stepper ist der einheitliche 44-px-Mengen-Wähler des Service-Bereichs. Plus
// ist dauerhaft primär, Minus outline; bei Menge 0 ist Minus regulär deaktiviert
// (abgeblendet und nicht antippbar wie jeder deaktivierte Outline-Button), damit
// der deaktivierte Zustand eindeutig ist. Die Menge in der Mitte
// hat feste Breite, damit der Zustandswechsel keinen Layout-Shift auslöst.
// addDisabled deckelt das Plus dort, wo eine Höchstmenge gilt (z. B. die
// unbezahlte Menge einer Position).
// Beide Tasten geben ein deutlich stärkeres Press-Feedback als der Standard-
// Button (scale .92 statt .99); die transform-only-Transition (100 ms linear)
// hält das Eindrücken knackig gemäß Motion-Inventar.
//
// minusNurAbEins blendet Minus und Mengenanzeige bei Menge 0 ganz aus, statt sie
// deaktiviert zu zeigen. Gedacht für die Bestellliste, wo jede Zeile nur eine
// Variante trägt: dort gingen die rund 80 px sonst dauerhaft von der Breite des
// Variantennamens ab, und ein gekürzter Name führt zur Fehlbuchung. Überall
// sonst bleibt der deaktivierte Minus-Knopf die klarere Anzeige.
export function Stepper({
  menge,
  onAdd,
  onRemove,
  addLabel,
  removeLabel,
  addDisabled = false,
  minusNurAbEins = false,
}: StepperProps) {
  const leer = menge === 0
  const minusZeigen = !minusNurAbEins || !leer

  return (
    <div className="flex items-center gap-2">
      {minusZeigen && (
        <>
          <Button
            size="icon"
            variant="outline"
            className="size-11 rounded-full transition-transform duration-100 ease-linear active:not-aria-[haspopup]:scale-[.92]"
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
        </>
      )}
      <Button
        size="icon"
        className="size-11 rounded-full transition-transform duration-100 ease-linear active:not-aria-[haspopup]:scale-[.92]"
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
