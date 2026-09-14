import { useEffect, useRef, useState } from 'react'

const DAUER_MS = 700

/**
 * Zählt einen ganzzahligen Wert bei Änderung über 700 ms animiert zum neuen
 * Wert und endet exakt am Zielwert; beim ersten Rendern wird nicht animiert.
 *
 * Ohne Animationsumgebung — reduzierte Bewegung, fehlendes
 * `requestAnimationFrame` oder `matchMedia` — erscheint sofort der Zielwert.
 * Das Test-Setup meldet reduzierte Bewegung; ein Test, der animieren will,
 * muss `matchMedia` selbst stubben.
 */
export function useCountUp(ziel: number): number {
  const [wert, setWert] = useState(ziel)
  const angezeigtRef = useRef(ziel)

  useEffect(() => {
    const von = angezeigtRef.current
    if (von === ziel) return
    if (!animierbar()) {
      // Der Zielwert muss im gerenderten Ergebnis stehen; das synchrone
      // setState ist hier gewollt, daher die Ausnahme von der Lint-Regel.
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

function animierbar(): boolean {
  return (
    typeof requestAnimationFrame === 'function' &&
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    !window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )
}
