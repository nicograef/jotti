import { useSyncExternalStore } from 'react'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

export function useAnzahlOffeneVorgaenge(): number {
  return useSyncExternalStore(
    VorgangsRegisterSingleton.abonnieren,
    VorgangsRegisterSingleton.anzahlOffen,
  )
}
