import { useEffect, useRef } from 'react'

/**
 * Meldet `true` genau beim ersten Rendern, in dem `bereit` gilt (dem „ersten
 * Aufbau" einer Liste), danach dauerhaft `false`. So animiert der
 * Listen-Eintritt nur beim ersten Aufbau und nie bei späteren Daten-Refetches.
 *
 * `bereit` überspringt das Skeleton-Vorspiel: Solange die Liste noch lädt
 * (`bereit === false`), zählt kein Aufbau; erst das erste Rendern mit Daten
 * löst den Eintritt aus.
 *
 * Das Flag liegt bewusst in einem Ref: Es darf kein zusätzliches Rendern
 * auslösen, sonst würde die frisch gestartete Animation abgerissen. Der Ref wird
 * ausschließlich im Effekt geschrieben und beim Rendern gelesen — Letzteres
 * meldet die Lint-Regel `react-hooks/refs`, hier ist es aber genau das gewollte
 * Verhalten (der Wert steuert nur die Animationsklasse, nicht die Darstellung).
 */
export function useErstAufbau(bereit: boolean): boolean {
  const aufgebautRef = useRef(false)
  useEffect(() => {
    if (bereit) {
      aufgebautRef.current = true
    }
  }, [bereit])
  // eslint-disable-next-line react-hooks/refs
  return bereit && !aufgebautRef.current
}
