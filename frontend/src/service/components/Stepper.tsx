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

// Einheitlicher 44-px-Mengen-Wähler des Service-Bereichs. Bei Menge 0 ist Minus
// regulär deaktiviert; die Menge in der Mitte hat feste Breite, damit der
// Zustandswechsel keinen Layout-Shift auslöst.
//
// minusNurAbEins blendet Minus und Mengenanzeige bei Menge 0 ganz aus: in der
// Bestellliste trägt jede Zeile eine Variante, ein deaktivierter Minus-Knopf je
// Zeile füllt die Liste. Überall sonst bleibt der deaktivierte Minus die klarere
// Anzeige. Der Aufrufort muss dafür die volle Stepper-Breite reservieren
// (ProductList: 8,25 rem), sonst wächst der Stepper beim ersten Tap und
// verschiebt Namensumbruch und Folgezeilen.
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
