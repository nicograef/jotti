import {
  type Kategorie,
  KATEGORIE_LABEL,
  KATEGORIE_ORDER,
  type Produkt,
  type Steuersatz,
} from '@/lib/produktSchemas'

import {
  type DruckstationConfig,
  STATION_KATEGORIE_LABEL,
} from '../settings/DruckstationBackend'
import { STEUERSATZ_LABEL } from './Produkt'

export interface ProduktGruppe {
  kategorie: Kategorie
  label: string
  produkte: Produkt[]
}

// Leere Kategorien entfallen; die Produktreihenfolge liefert das Backend
// bereits sortiert.
export function groupProdukteByKategorie(produkte: Produkt[]): ProduktGruppe[] {
  return KATEGORIE_ORDER.map((kategorie) => ({
    kategorie,
    label: KATEGORIE_LABEL[kategorie],
    produkte: produkte.filter((p) => p.kategorie === kategorie),
  })).filter((gruppe) => gruppe.produkte.length > 0)
}

// null, sobald die Gruppe uneinheitlich ist — lieber kein Steuersatz-Zusatz als
// eine falsche Sammelangabe.
export function gemeinsamerSteuersatz(produkte: Produkt[]): Steuersatz | null {
  if (produkte.length === 0) {
    return null
  }
  const erster = produkte[0].steuersatz
  return produkte.every((p) => p.steuersatz === erster) ? erster : null
}

export function kategorieZusatz(
  kategorie: Kategorie,
  produkte: Produkt[],
  druckstationen: DruckstationConfig[],
): string {
  const teile: string[] = []

  const steuersatz = gemeinsamerSteuersatz(produkte)
  if (steuersatz !== null) {
    teile.push(STEUERSATZ_LABEL[steuersatz])
  }

  const station = druckstationen.find((s) => s.kategorie === kategorie)
  if (station && station.druckerIp !== '') {
    teile.push(`Bons an Station „${STATION_KATEGORIE_LABEL[kategorie]}"`)
  }

  return teile.join(' · ')
}

// Nur nicht gelöschte Varianten zählen (Backend liefert bereits gefiltert).
export function produktUnterzeile(produkte: Produkt[]): string {
  const anzahlProdukte = produkte.length
  const anzahlVarianten = produkte.reduce(
    (summe, p) => summe + p.varianten.length,
    0,
  )
  const produkteText = `${String(anzahlProdukte)} Produkt${anzahlProdukte === 1 ? '' : 'e'}`
  const variantenText = `${String(anzahlVarianten)} Variante${anzahlVarianten === 1 ? '' : 'n'}`
  return `${produkteText} · ${variantenText} · Änderungen wirken sofort auf allen Service-Handys`
}
