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

// Eine rueckgabeId je Vorgang: `vorgangKey` fasst Modus und Auswahl der Zahlung
// oder des Verkaufs, Pfand-Mengen behalten die ID. Neu wird sie bei neuem
// Vorgang und bei leerer Pfand-Auswahl, also auch nach Erfolg; eine abweichende
// Wiederholung weist der Server mit `pfandrueckgabe_abweichend` ab.
export function usePfandAuswahl(
  pfandProdukte: Produkt[],
  vorgangKey: string,
): PfandAuswahl {
  const steuerung = useMengen<number>()
  const { receiptItems, inputItems } = toBestellungData(
    pfandProdukte,
    steuerung.mengen,
  )

  const schluessel = `${String(inputItems.length === 0)}|${vorgangKey}`
  const [vorgang, setVorgang] = useState(() => ({
    schluessel,
    rueckgabeId: crypto.randomUUID(),
  }))
  if (vorgang.schluessel !== schluessel) {
    setVorgang({ schluessel, rueckgabeId: crypto.randomUUID() })
  }

  return {
    steuerung,
    positionen: inputItems,
    summeCents: calculateTotalPrice(receiptItems),
    rueckgabeId: vorgang.rueckgabeId,
  }
}
