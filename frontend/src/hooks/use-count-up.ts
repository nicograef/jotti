import { useEffect, useRef, useState } from 'react'

const DAUER_MS = 700

/**
 * Zählt einen ganzzahligen Wert (z. B. Cent-Beträge) bei Änderung über 700 ms
 * animiert vom alten zum neuen Wert (ease-out-cubic `1-(1-p)^3`, per
 * `requestAnimationFrame`) und endet exakt am Zielwert. Beim ersten Rendern
 * wird nicht animiert — der Startwert steht bereits.
 *
 * Ohne Animationsumgebung — reduzierte Bewegung, fehlendes
 * `requestAnimationFrame` oder `matchMedia` (jsdom/Testumgebung) — erscheint
 * sofort der Zielwert. In jsdom hat rAF keine echte Zeitbasis; Tests prüfen
 * deshalb den Endzustand, nicht die Zwischenwerte.
 */
export function useCountUp(ziel: number): number {
  const [wert, setWert] = useState(ziel)
  // Zuletzt angezeigter Wert: Startpunkt der nächsten Animation und Grundlage
  // der Änderungserkennung. Liegt in einem Ref, um kein zusätzliches Rendern
  // auszulösen (der Ref wird nur im Effekt gelesen und geschrieben).
  const angezeigtRef = useRef(ziel)

  useEffect(() => {
    const von = angezeigtRef.current
    // Kein Wechsel (u. a. erstes Rendern): nichts zu animieren.
    if (von === ziel) return
    if (!animierbar()) {
      // Ohne Animation direkt auf den Zielwert springen. Das synchrone setState
      // ist hier gewollt und unvermeidbar (das gerenderte Ergebnis muss den
      // neuen Wert zeigen), daher die Ausnahme von der set-state-in-effect-Regel.
      angezeigtRef.current = ziel
      // eslint-disable-next-line react-x/set-state-in-effect
      setWert(ziel)
      return
    }
    const t0 = performance.now()
    let frame = requestAnimationFrame(function schritt(t) {
      const p = Math.min(1, (t - t0) / DAUER_MS)
      const e = 1 - Math.pow(1 - p, 3)
      const aktuell = Math.round(von + (ziel - von) * e)
      angezeigtRef.current = aktuell
      setWert(aktuell)
      if (p < 1) frame = requestAnimationFrame(schritt)
    })
    return () => {
      cancelAnimationFrame(frame)
    }
  }, [ziel])

  return wert
}

/**
 * Prüft, ob animiert werden darf: nur mit echter rAF-Zeitbasis und ohne die
 * Nutzerpräferenz für reduzierte Bewegung. Fehlt `matchMedia` (jsdom), gilt die
 * Umgebung als nicht animierbar, sodass der Hook sofort den Zielwert liefert.
 */
function animierbar(): boolean {
  return (
    typeof requestAnimationFrame === 'function' &&
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    !window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )
}
