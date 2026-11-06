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

// Die rueckgabeId gehört zu genau einem Vorgang aus Pfand-Auswahl und
// `vorgangKey`, der Auswahl der Zahlung oder des Verkaufs. Jede Änderung
// erneuert sie, nur ein unveränderter Wiederholversuch behält sie; sonst meldet
// der Server nach einem Timeout Erfolg für einen anderen Vorgang.
export function usePfandAuswahl(
  pfandProdukte: Produkt[],
  vorgangKey: string,
): PfandAuswahl {
  const steuerung = useMengen<number>()
  const { receiptItems, inputItems } = toBestellungData(
    pfandProdukte,
    steuerung.mengen,
  )

  const auswahlKey = `${inputItems
    .map((p) => `${p.varianteId.toString()}:${p.menge.toString()}`)
    .join(',')}|${vorgangKey}`
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
