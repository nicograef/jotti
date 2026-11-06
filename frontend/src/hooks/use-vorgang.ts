import { useState } from 'react'

import { type MengenSteuerung, useMengen } from './use-mengen'

// Eine Auswahl samt Idempotenz-Schlüssel des logischen Vorgangs (D08). Die Seite
// hält beide, damit sie das Aus- und Einhängen der Tab-Inhalte überstehen.
export interface Vorgang<K extends string | number> {
  steuerung: MengenSteuerung<K>
  // Neu, sobald die Auswahl leer wird oder sich wieder füllt; ein Retry behält ihn.
  vorgangId: string
  // Schlüssel für einen Absendeversuch mit diesem Inhalt. Nach einem Fehlversuch
  // mit anderem Inhalt ist der alte verbraucht: das Backend hielte die geänderte
  // Buchung für ein Duplikat und verwürfe sie stillschweigend.
  schluesselFuer: (inhalt: unknown) => string
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
  }))
  if (stand.leer !== leer) {
    setStand({ leer, vorgangId: crypto.randomUUID(), letzterVersuch: null })
  }

  const schluesselFuer = (inhalt: unknown): string => {
    const versuch = JSON.stringify(inhalt)
    const vorgangId =
      stand.letzterVersuch === null || stand.letzterVersuch === versuch
        ? stand.vorgangId
        : crypto.randomUUID()
    setStand({ ...stand, vorgangId, letzterVersuch: versuch })
    return vorgangId
  }

  return { steuerung, vorgangId: stand.vorgangId, schluesselFuer }
}
