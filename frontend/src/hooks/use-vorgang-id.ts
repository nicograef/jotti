import { useState } from 'react'

// Vergibt den Idempotenz-Schlüssel eines fachlichen Vorgangs: Ein Schlüssel gilt
// für eine ganze Zusammenstellung — von der ersten Auswahl bis zur Buchung — und
// wird erst abgelöst, wenn nach dem Leerzustand eine neue Zusammenstellung
// beginnt. Ein Wiederholversuch behält den Schlüssel deshalb immer, auch wenn
// die Auswahl inzwischen eine andere ist.
//
// Bewusst NICHT an die Nutzdaten gebunden: Der Server bindet den Schlüssel an
// die Nutzdaten und meldet eine Zweiteinreichung mit abweichender Auswahl
// ausdrücklich (409 `vorgang_daten_abweichend`). Rotierte der Client bei jeder
// Änderung, sähe der Server denselben Schlüssel nie zweimal — die Prüfung liefe
// ins Leere, und der gefährliche Fall bliebe unerkannt: Die erste Einreichung
// ist gebucht, nur ihre Antwort ging verloren; der Helfer ergänzt eine Position
// und sendet erneut. Mit rotiertem Schlüssel bucht der Server ein zweites Mal,
// mit stabilem Schlüssel beanstandet er die Abweichung.
//
// Der Abgleich läuft im Render (React-idiomatischer State-Sync wie beim
// Tischwechsel in TablePage), nicht in einem useEffect: Der Schlüssel muss
// synchron vor dem nächsten Submit feststehen.
//
// `istLeer` ist der fachliche Leerzustand der Zusammenstellung an der
// Aufrufstelle — in aller Regel „keine Position ausgewählt".
//
// Der Hook gehört dorthin, wo die Zusammenstellung selbst liegt. Wird er
// unterhalb einer Grenze aufgerufen, die aus- und wieder eingehängt wird (etwa
// im Inhalt eines Radix-Tabs), bekommt eine unveränderte Auswahl nach dem
// Wiedereinhängen einen neuen Schlüssel — aus einem Wiederholversuch würde eine
// zweite Buchung.
export function useVorgangId(istLeer: boolean): string {
  const [vorgang, setVorgang] = useState(() => ({
    warLeer: istLeer,
    id: crypto.randomUUID(),
  }))

  if (vorgang.warLeer === istLeer) {
    return vorgang.id
  }

  // Nur der Übergang leer → nicht leer beginnt einen neuen Vorgang. Der Rückweg
  // merkt sich bloß den Leerzustand, damit die nächste Zusammenstellung wieder
  // rotiert. Bewusst zufällig statt aus den Nutzdaten abgeleitet: Zwei getrennte
  // Vorgänge mit identischen Nutzdaten (dieselbe Bestellung zweimal
  // nacheinander) brauchen verschiedene Schlüssel. Konvergiert nach einem
  // Re-Render.
  // eslint-disable-next-line react-x/purity
  const id = vorgang.warLeer ? crypto.randomUUID() : vorgang.id
  setVorgang({ warLeer: istLeer, id })
  return id
}
