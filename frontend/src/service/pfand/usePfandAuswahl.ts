import { useState } from 'react'

import { type MengenSteuerung, useMengen } from '@/hooks/use-mengen'
import type { Produkt } from '@/lib/produktSchemas'

import type { VerkaufPositionInput } from '../direktverkauf/Direktverkauf'
import { calculateTotalPrice, toBestellungData } from '../table/drawerUtils'

export interface PfandAuswahl {
  steuerung: MengenSteuerung<number>
  positionen: VerkaufPositionInput[]
  summeCents: number
  rueckgabeId: string
}

// Die rueckgabeId gehört zu genau einer Auswahl: Jede Änderung erneuert sie,
// nur ein unveränderter Wiederholversuch behält sie. Sonst bucht der Server
// nach einem Timeout die alte Menge, während die neue bar ausgezahlt wird.
export function usePfandAuswahl(pfandProdukte: Produkt[]): PfandAuswahl {
  const steuerung = useMengen<number>()
  const { receiptItems, inputItems } = toBestellungData(
    pfandProdukte,
    steuerung.mengen,
  )

  const auswahlKey = inputItems
    .map((p) => `${p.varianteId.toString()}:${p.menge.toString()}`)
    .join(',')
  const [vorgang, setVorgang] = useState(() => ({
    auswahlKey,
    rueckgabeId: crypto.randomUUID(),
  }))
  if (vorgang.auswahlKey !== auswahlKey) {
    setVorgang({ auswahlKey, rueckgabeId: crypto.randomUUID() })
  }

  return {
    steuerung,
    positionen: inputItems,
    summeCents: calculateTotalPrice(receiptItems),
    rueckgabeId: vorgang.rueckgabeId,
  }
}
