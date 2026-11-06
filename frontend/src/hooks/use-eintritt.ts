import { type CSSProperties, useState } from 'react'

const STAFFELUNG_MS = 60

// Animierter Eintritt an Position `index` einer Staffelung; `undefined` heißt
// ohne Animation. Beim Mount erfasst, damit ein Refetch die Staffelung nicht
// mittendrin abreißt.
export function useEintritt(index: number | undefined): {
  className?: string
  style?: CSSProperties
} {
  const [eintritt] = useState(index)
  if (eintritt === undefined) return {}
  return {
    className: 'animate-eintritt',
    style: { animationDelay: `${(eintritt * STAFFELUNG_MS).toString()}ms` },
  }
}
