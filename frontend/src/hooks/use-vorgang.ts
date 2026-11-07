import { useState } from 'react'

import { type MengenSteuerung, useMengen } from './use-mengen'

// Eine Auswahl samt Kommentaren und Idempotenz-Schlüssel des logischen Vorgangs
// (D08). Die Seite hält alles, damit es das Aus- und Einhängen der
// Tab-Inhalte übersteht: ein Retry trägt so denselben Inhalt und Schlüssel.
export interface Vorgang<K extends string | number> {
  steuerung: MengenSteuerung<K>
  // Neu, sobald die Auswahl leer wird oder sich wieder füllt; ein Retry behält ihn.
  vorgangId: string
  // Schlüssel für einen Absendeversuch mit diesem Inhalt. Nach einem Fehlversuch
  // mit anderem Inhalt ist der alte verbraucht: das Backend verwürfe die geänderte
  // Buchung stillschweigend als Duplikat (Bestellung, Direktverkauf) oder wiese
  // sie mit vorgang_abweichend ab (Zahlung).
  schluesselFuer: (inhalt: unknown) => string
  // Leer, sobald die Auswahl leer wird oder sich wieder füllt.
  kommentar: string
  setKommentar: (kommentar: string) => void
  // Je Schlüssel der Auswahl; fällt dessen Menge auf null, ist sein Kommentar weg.
  positionsKommentare: Partial<Record<K, string>>
  setPositionsKommentar: (key: K, kommentar: string) => void
}

export function useVorgang<K extends string | number>(
  max?: (key: K) => number,
): Vorgang<K> {
  const steuerung = useMengen<K>(max)
  const mengen: Partial<Record<string | number, number>> = steuerung.mengen
  const leer = !Object.values<number>(steuerung.mengen).some(
    (menge) => menge > 0,
  )
  const [stand, setStand] = useState(() => ({
    leer,
    vorgangId: crypto.randomUUID(),
    letzterVersuch: null as string | null,
    kommentar: '',
    positionsKommentare: {} as Partial<Record<K, string>>,
  }))
  const verwaist = Object.keys(stand.positionsKommentare).filter(
    (key) => !((mengen[key] ?? 0) > 0),
  )
  if (stand.leer !== leer) {
    setStand({
      leer,
      vorgangId: crypto.randomUUID(),
      letzterVersuch: null,
      kommentar: '',
      positionsKommentare: {},
    })
  } else if (verwaist.length > 0) {
    setStand((vorher) => ({
      ...vorher,
      positionsKommentare: ohneSchluessel(vorher.positionsKommentare, verwaist),
    }))
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

  const setPositionsKommentar = (key: K, kommentar: string) => {
    setStand((vorher) => ({
      ...vorher,
      positionsKommentare: { ...vorher.positionsKommentare, [key]: kommentar },
    }))
  }

  return {
    steuerung,
    vorgangId: stand.vorgangId,
    schluesselFuer,
    kommentar: stand.kommentar,
    setKommentar,
    positionsKommentare: stand.positionsKommentare,
    setPositionsKommentar,
  }
}

function ohneSchluessel<T extends object>(objekt: T, keys: string[]): T {
  return Object.fromEntries(
    Object.entries(objekt).filter(([key]) => !keys.includes(key)),
  ) as T
}
