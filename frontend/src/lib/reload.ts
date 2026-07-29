/**
 * Lädt die Seite neu — die eine Stelle, an der der Versions-Handshake das tut.
 *
 * Eigene Funktion und kein direkter Aufruf am Aufrufort, weil jsdom
 * `window.location.reload` nicht ersetzen lässt („Cannot redefine property").
 * Nur über ein eigenes Modul lässt sich der erzwungene Reload im Test
 * beobachten, statt ihn bloß als „Not implemented: navigation" zu sehen.
 */
export function seiteNeuLaden(): void {
  window.location.reload()
}
