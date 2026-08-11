import type { DruckstationConfig } from '../settings/DruckstationBackend'
import {
  Kategorie,
  type Produkt,
  type Steuersatz,
  STEUERSATZ_LABEL,
} from './Produkt'

// Feste Anzeigereihenfolge der Kategorie-Abschnitte in der Preisliste.
export const KATEGORIE_ORDER: Kategorie[] = [
  Kategorie.ESSEN,
  Kategorie.GETRAENK,
  Kategorie.SONSTIGES,
]

// Deutsche Abschnitts-Überschriften je Kategorie (Design-Handoff 1c).
export const KATEGORIE_LABEL: Record<Kategorie, string> = {
  essen: 'Essen',
  getraenk: 'Getränke',
  sonstiges: 'Sonstiges',
}

// Stationsname je Kategorie für den Zusatz „Bons an Station …". Die
// Druckstation trägt dieselbe Kategorie wie das Produkt; nur der Anzeigename
// weicht leicht ab (Singular „Getränk").
const KATEGORIE_STATION_LABEL: Record<Kategorie, string> = {
  essen: 'Essen',
  getraenk: 'Getränk',
  sonstiges: 'Sonstiges',
}

export interface ProduktGruppe {
  kategorie: Kategorie
  label: string
  produkte: Produkt[]
}

// Gruppiert die Produkte in der festen Kategorie-Reihenfolge. Leere Kategorien
// werden ausgelassen; die Produktreihenfolge innerhalb einer Gruppe bleibt
// erhalten (Backend liefert nach Reihenfolge sortiert).
export function groupProdukteByKategorie(produkte: Produkt[]): ProduktGruppe[] {
  return KATEGORIE_ORDER.map((kategorie) => ({
    kategorie,
    label: KATEGORIE_LABEL[kategorie],
    produkte: produkte.filter((p) => p.kategorie === kategorie),
  })).filter((gruppe) => gruppe.produkte.length > 0)
}

// Einheitlicher Steuersatz einer Gruppe, sofern alle Produkte darin denselben
// tragen — sonst null (dann wird der Steuersatz-Zusatz weggelassen, statt eine
// falsche Sammelangabe zu machen).
export function gemeinsamerSteuersatz(produkte: Produkt[]): Steuersatz | null {
  if (produkte.length === 0) {
    return null
  }
  const erster = produkte[0].steuersatz
  return produkte.every((p) => p.steuersatz === erster) ? erster : null
}

// Zusatzzeile hinter der Kategorie-Überschrift: Steuersatz (wenn einheitlich)
// und Stations-Hinweis (wenn für diese Kategorie ein Drucker konfiguriert ist).
// Beide Teile sind optional; fehlen beide, ist die Zusatzzeile leer.
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
    teile.push(`Bons an Station „${KATEGORIE_STATION_LABEL[kategorie]}"`)
  }

  return teile.join(' · ')
}

// Kopf-Unterzeile: „n Produkte · m Varianten · Änderungen wirken sofort …".
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
