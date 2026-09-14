import { useEffect } from 'react'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

/**
 * Der einzige Weg, ins Vorgangs-Register zu schreiben: meldet einen offenen
 * Vorgang, solange `offen` gilt, und gibt ihn im Effekt-Cleanup wieder frei.
 *
 * Das Cleanup ist zwingend: Ein Tischwechsel setzt nur Zustand zurück (die
 * Seite bleibt gemountet), der Wechsel zwischen Drawer- und Spaltenlayout
 * tauscht ganze Teilbäume aus. Eine stehen gebliebene Anmeldung ließe den
 * erzwungenen Reload für immer auf einen Vorgang warten, den es nicht gibt.
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
