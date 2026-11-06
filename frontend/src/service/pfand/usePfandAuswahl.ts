import { useEffect, useRef, useState } from 'react'

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

// Eine rueckgabeId je Vorgang: Sie wechselt, wenn eine Auswahl aus dem Leeren
// beginnt; ein Wiederholversuch behält sie. Solange `gesperrt` gilt, wird eine
// Auswahl im Render verworfen.
export function usePfandAuswahl(
  pfandProdukte: Produkt[],
  gesperrt: boolean,
): PfandAuswahl {
  const steuerung = useMengen<number>()
  const { receiptItems, inputItems } = toBestellungData(
    pfandProdukte,
    steuerung.mengen,
  )
  const leer = inputItems.length === 0
  if (gesperrt && !leer) {
    steuerung.reset()
  }

  const [rueckgabeId, setRueckgabeId] = useState(() => crypto.randomUUID())
  const warLeerRef = useRef(leer)
  useEffect(() => {
    if (warLeerRef.current && !leer) {
      setRueckgabeId(crypto.randomUUID())
    }
    warLeerRef.current = leer
  }, [leer])

  return {
    steuerung,
    positionen: inputItems,
    summeCents: calculateTotalPrice(receiptItems),
    rueckgabeId,
  }
}
