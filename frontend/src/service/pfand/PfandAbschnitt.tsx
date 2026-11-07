import type { MengenSteuerung } from '@/hooks/use-mengen'
import type { Produkt } from '@/lib/produktSchemas'
import { formatEuro } from '@/lib/utils'

import { ProductList } from '../ProductList'

interface PfandAbschnittProps {
  pfandProdukte: Produkt[]
  steuerung: MengenSteuerung<number>
  summeCents: number
  disabled: boolean
}

// Das deaktivierte fieldset sperrt alle Stepper-Buttons der Liste auf einmal.
export function PfandAbschnitt(props: PfandAbschnittProps) {
  return (
    <section aria-labelledby="pfand-zurueck" className="px-4 pt-3">
      <div className="flex items-baseline justify-between gap-3">
        <h3 id="pfand-zurueck" className="text-sm font-semibold">
          Pfand zurück
        </h3>
        {props.summeCents > 0 && (
          <span className="text-sm font-semibold tabular-nums">
            {formatEuro(-props.summeCents)}
          </span>
        )}
      </div>
      <fieldset disabled={props.disabled} className="disabled:opacity-50">
        <ProductList
          products={props.pfandProdukte}
          variantMengen={props.steuerung.mengen}
          onAdd={props.steuerung.add}
          onRemove={props.steuerung.remove}
        />
      </fieldset>
    </section>
  )
}
