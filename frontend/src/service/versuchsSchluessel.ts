// Idempotenz-Schlüssel der Storno- und Umbuchungsversuche aus einer Historie.
// Ein Versuch mit demselben Inhalt behält seinen Schlüssel, bis einer gelingt:
// eine verlorene Antwort bucht so nicht doppelt, auch nach Schließen und
// Wiederöffnen des Drawers. Anderer Inhalt bekommt einen eigenen Schlüssel.
const offeneVersuche = new Map<string, string>()

export async function mitVersuchsSchluessel(
  inhalt: unknown,
  senden: (schluessel: string) => Promise<void>,
): Promise<void> {
  const json = JSON.stringify(inhalt)
  let schluessel = offeneVersuche.get(json)
  if (schluessel === undefined) {
    schluessel = crypto.randomUUID()
    offeneVersuche.set(json, schluessel)
  }
  await senden(schluessel)
  offeneVersuche.delete(json)
}
