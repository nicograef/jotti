import { useState } from 'react'

import { useOffenerVorgang } from './use-offener-vorgang'

export interface MengenSteuerung<K extends string | number> {
  mengen: Record<K, number>
  add: (key: K) => void
  remove: (key: K) => void
  reset: () => void
  setAll: (next: Record<K, number>) => void
}

/**
 * Quantity-selector state keyed by id. `setAll` bypasses `max` — the caller
 * must stay within the cap. Pass `max` to cap a key's quantity on `add`.
 */
export function useMengen<K extends string | number>(
  max?: (key: K) => number,
): MengenSteuerung<K> {
  const [mengen, setMengen] = useState<Record<K, number>>(
    () => ({}) as Record<K, number>,
  )

  // Any selected quantity is work that a forced reload would throw away.
  // Reported centrally here rather than at the six call sites.
  useOffenerVorgang(Object.values<number>(mengen).some((menge) => menge > 0))

  const add = (key: K) => {
    setMengen((prev) => {
      const current = prev[key] || 0
      if (max && current >= max(key)) return prev
      return { ...prev, [key]: current + 1 }
    })
  }

  const remove = (key: K) => {
    setMengen((prev) => {
      const current = prev[key] || 0
      if (current <= 0) return prev
      return { ...prev, [key]: current - 1 }
    })
  }

  const reset = () => {
    setMengen({} as Record<K, number>)
  }

  const setAll = (next: Record<K, number>) => {
    setMengen(next)
  }

  return { mengen, add, remove, reset, setAll }
}
