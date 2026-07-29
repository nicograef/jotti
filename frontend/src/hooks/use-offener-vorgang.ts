import { useEffect } from 'react'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

/**
 * Meldet einen offenen Vorgang, solange `offen` gilt, und gibt ihn wieder frei,
 * sobald `offen` entfällt oder die meldende Komponente ausgehängt wird. Der
 * einzige Weg, ins Vorgangs-Register zu schreiben.
 *
 * Die Abmeldung gehört zwingend in den Effekt-Cleanup: Ein Tischwechsel setzt
 * nur Zustand zurück (die Seite bleibt gemountet), und der Wechsel zwischen
 * Drawer- und Spaltenlayout tauscht ganze Teilbäume aus — beides ohne Zutun der
 * meldenden Stelle. Bliebe eine Anmeldung dabei stehen, wartete der erzwungene
 * Reload für immer auf einen Vorgang, den es nicht mehr gibt.
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
