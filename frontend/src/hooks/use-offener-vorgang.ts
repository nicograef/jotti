import { useEffect } from 'react'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

/**
 * The only way to write to the Vorgangs-Register: registers an open Vorgang while `offen` holds and releases it in the effect cleanup.
 * The cleanup is mandatory: a Tisch switch only resets state and a layout switch swaps subtrees, so a stale entry would block the forced reload forever.
 */
export function useOffenerVorgang(offen: boolean): void {
  useEffect(() => {
    if (!offen) return
    VorgangsRegisterSingleton.anmelden()
    return () => {
      VorgangsRegisterSingleton.abmelden()
    }
  }, [offen])
}
