import { useState } from 'react'

import { type MengenSteuerung, useMengen } from '@/hooks/use-mengen'
import { BackendError } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'

import type { VerkaufPositionInput } from '../direktverkauf/Direktverkauf'
import { calculateTotalPrice, toBestellungData } from '../table/drawerUtils'

// Die Seite hält den Vorgang, damit Mengen und rueckgabeId Tab-Wechsel überstehen.
export interface PfandVorgang {
  steuerung: MengenSteuerung<number>
  rueckgabeId: string
}

export interface PfandAuswahl extends PfandVorgang {
  positionen: VerkaufPositionInput[]
  summeCents: number
}

// Neu wird die rueckgabeId nur mit leerer Pfand-Auswahl, also nach Erfolg und Reset; Mengen und
// Zahlungsauswahl behalten sie. Eine gebuchte Wiederholung mit anderem Inhalt weist der Server
// mit `pfandrueckgabe_abweichend` ab.
export function usePfandVorgang(): PfandVorgang {
  const steuerung = useMengen<number>()
  const leer = !Object.values<number>(steuerung.mengen).some(
    (menge) => menge > 0,
  )
  const [vorgang, setVorgang] = useState(() => ({
    leer,
    rueckgabeId: crypto.randomUUID(),
  }))
  if (vorgang.leer !== leer) {
    setVorgang({ leer, rueckgabeId: crypto.randomUUID() })
  }

  return { steuerung, rueckgabeId: vorgang.rueckgabeId }
}

export function pfandAuswahl(
  pfandProdukte: Produkt[],
  vorgang: PfandVorgang,
): PfandAuswahl {
  const { receiptItems, inputItems } = toBestellungData(
    pfandProdukte,
    vorgang.steuerung.mengen,
  )
  return {
    ...vorgang,
    positionen: inputItems,
    summeCents: calculateTotalPrice(receiptItems),
  }
}

// Nach `pfandrueckgabe_abweichend` leert der Abschluss die Pfand-Auswahl: erst das gibt der neu
// erfassten Rückgabe eine neue rueckgabeId.
export function istPfandrueckgabeAbweichend(error: unknown): boolean {
  return (
    error instanceof BackendError && error.code === 'pfandrueckgabe_abweichend'
  )
}
