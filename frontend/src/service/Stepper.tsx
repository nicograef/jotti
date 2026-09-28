import { Minus, Plus } from 'lucide-react'

import { Button } from '@/components/ui/button'

interface StepperProps {
  menge: number
  onAdd: () => void
  onRemove: () => void
  addLabel: string
  removeLabel: string
  addDisabled?: boolean
  // Hides Minus and quantity at 0, where a disabled Minus in every row would crowd the order list.
  // The caller must reserve the full Stepper width (ProductList: 8.25 rem), else the first tap shifts the rows.
  minusNurAbEins?: boolean
}

// Service-wide 44 px quantity picker; the centre quantity has a fixed width so state changes cause no layout shift.
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
