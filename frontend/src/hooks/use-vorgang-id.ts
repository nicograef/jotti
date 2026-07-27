import { useState } from 'react'

// Bindet den Idempotenz-Schlüssel eines fachlichen Vorgangs an dessen
// Nutzdaten: gleiche Nutzdaten = derselbe Vorgang = derselbe Schlüssel;
// geänderte Nutzdaten (Positionen, Mengen, Kommentar, Ziel-Tisch, …) = neuer
// Vorgang = neuer Schlüssel. So bleibt ein echter Wiederholversuch idempotent,
// während eine nach einem Fehlversuch geänderte Auswahl nie unter dem alten
// Schlüssel gesendet wird — der Server würde sie sonst als Duplikat
// verschlucken und fälschlich Erfolg melden, ohne zu buchen.
//
// Der Abgleich läuft im Render (React-idiomatischer State-Sync wie beim
// Tischwechsel in TablePage), nicht in einem useEffect: Ein Submit zwischen
// Nutzdaten-Wechsel und Effekt könnte sonst den alten Schlüssel mit den neuen
// Nutzdaten senden.
//
// `nutzdaten` muss JSON-serialisierbar sein und pro Render mit stabiler
// Feld-Reihenfolge aufgebaut werden (Objekt-Literal an der Aufrufstelle,
// abgeleitet aus dem, was tatsächlich gesendet wird).
export function useVorgangId(nutzdaten: unknown): string {
  const nutzdatenJson = JSON.stringify(nutzdaten)
  const [vorgang, setVorgang] = useState(() => ({
    nutzdatenJson,
    id: crypto.randomUUID(),
  }))
  if (vorgang.nutzdatenJson !== nutzdatenJson) {
    // Bewusst zufällig im Render: Der Schlüssel muss synchron vor dem nächsten
    // Submit feststehen (useEffect wäre zu spät), und er darf nicht rein aus
    // den Nutzdaten abgeleitet sein — zwei getrennte Vorgänge mit identischen
    // Nutzdaten (gleiche Bestellung zweimal nacheinander) brauchen
    // verschiedene Schlüssel. Konvergiert nach einem Re-Render.
    // eslint-disable-next-line react-x/purity
    const neuerVorgang = { nutzdatenJson, id: crypto.randomUUID() }
    setVorgang(neuerVorgang)
    return neuerVorgang.id
  }
  return vorgang.id
}
