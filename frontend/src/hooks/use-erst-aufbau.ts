import { useEffect, useRef } from 'react'

/**
 * Meldet `true` genau beim ersten Rendern, in dem `bereit` gilt, danach
 * dauerhaft `false`. So animiert der Listen-Eintritt nur beim ersten Aufbau und
 * nie bei einem späteren Refetch; `bereit === false` überspringt das Skeleton.
 *
 * Das Flag liegt in einem Ref: Ein zusätzliches Rendern risse die frisch
 * gestartete Animation ab. Das Lesen beim Rendern meldet `react-hooks/refs`,
 * ist hier aber gewollt — der Wert steuert nur die Animationsklasse.
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
