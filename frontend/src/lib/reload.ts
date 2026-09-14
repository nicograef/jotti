/**
 * Eigene Funktion statt `window.location.reload()` am Aufrufort, weil jsdom
 * die Methode nicht ersetzen lässt („Cannot redefine property"). Nur über ein
 * eigenes Modul ist der erzwungene Reload im Test beobachtbar.
 */
export function seiteNeuLaden(): void {
  window.location.reload()
}
