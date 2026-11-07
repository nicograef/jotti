import type { MengenSteuerung } from '@/hooks/use-mengen'
import type { Vorgang } from '@/hooks/use-vorgang'
import { BackendError } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'

import { calculateTotalPrice, toBestellungData } from '../drawerUtils'
import type { PositionInput } from '../schemas'

// Die Seite hält den Pfand-Vorgang; seine vorgangId ist die rueckgabeId. Eine
// gebuchte Wiederholung mit anderem Inhalt weist der Server mit
// `pfandrueckgabe_abweichend` ab.
export type PfandVorgang = Vorgang<number>

export interface PfandAuswahl {
  steuerung: MengenSteuerung<number>
  rueckgabeId: string
  positionen: PositionInput[]
  summeCents: number
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
    steuerung: vorgang.steuerung,
    rueckgabeId: vorgang.vorgangId,
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
