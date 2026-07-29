import { useSyncExternalStore } from 'react'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

/**
 * Liefert die Anzahl der gerade offenen Vorgänge und rendert bei jeder Änderung
 * neu. `useSyncExternalStore` ist die React-Schnittstelle für Zustand außerhalb
 * von React — kein Abo von Hand, kein zusätzlicher Context.
 */
export function useAnzahlOffeneVorgaenge(): number {
  return useSyncExternalStore(
    VorgangsRegisterSingleton.abonnieren,
    VorgangsRegisterSingleton.anzahlOffen,
  )
}
