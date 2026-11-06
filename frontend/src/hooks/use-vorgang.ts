import { useState } from 'react'

import { type MengenSteuerung, useMengen } from './use-mengen'

// Eine Auswahl samt Kommentar und Idempotenz-Schlüssel des logischen Vorgangs
// (D08). Die Seite hält alle drei, damit sie das Aus- und Einhängen der
// Tab-Inhalte überstehen: ein Retry trägt so denselben Inhalt und Schlüssel.
export interface Vorgang<K extends string | number> {
  steuerung: MengenSteuerung<K>
  // Neu, sobald die Auswahl leer wird oder sich wieder füllt; ein Retry behält ihn.
  vorgangId: string
  // Schlüssel für einen Absendeversuch mit diesem Inhalt. Nach einem Fehlversuch
  // mit anderem Inhalt ist der alte verbraucht: das Backend hielte die geänderte
  // Buchung für ein Duplikat und verwürfe sie stillschweigend.
  schluesselFuer: (inhalt: unknown) => string
  // Leer, sobald die Auswahl leer wird oder sich wieder füllt.
  kommentar: string
  setKommentar: (kommentar: string) => void
}

export function useVorgang<K extends string | number>(
  max?: (key: K) => number,
): Vorgang<K> {
  const steuerung = useMengen<K>(max)
  const leer = !Object.values<number>(steuerung.mengen).some(
    (menge) => menge > 0,
  )
  const [stand, setStand] = useState(() => ({
    leer,
    vorgangId: crypto.randomUUID(),
    letzterVersuch: null as string | null,
    kommentar: '',
  }))
  if (stand.leer !== leer) {
    setStand({
      leer,
      vorgangId: crypto.randomUUID(),
      letzterVersuch: null,
      kommentar: '',
    })
  }

  const schluesselFuer = (inhalt: unknown): string => {
    const versuch = JSON.stringify(inhalt)
    const vorgangId =
      stand.letzterVersuch === null || stand.letzterVersuch === versuch
        ? stand.vorgangId
        : crypto.randomUUID()
    setStand((vorher) => ({ ...vorher, vorgangId, letzterVersuch: versuch }))
    return vorgangId
  }

  const setKommentar = (kommentar: string) => {
    setStand((vorher) => ({ ...vorher, kommentar }))
  }

  return {
    steuerung,
    vorgangId: stand.vorgangId,
    schluesselFuer,
    kommentar: stand.kommentar,
    setKommentar,
  }
}
