import { useState } from 'react'

/**
 * Return shape of {@link useMengen}. Exposed so callers can lift the selection
 * state and pass the whole controller down as a single prop (e.g. TablePage
 * hoists the order cart and payment selection so they survive tab switches).
 */
export interface MengenSteuerung<K extends string | number> {
  mengen: Record<K, number>
  add: (key: K) => void
  remove: (key: K) => void
  reset: () => void
  setAll: (next: Record<K, number>) => void
}

/**
 * Quantity-selector state keyed by id (variant id for ordering/direct sale,
 * position id for payment). `add` increments, `remove` decrements but never
 * below zero, `reset` clears the whole selection, `setAll` replaces the whole
 * selection with the given quantities (without applying `max` — the caller is
 * responsible for staying within the cap). Pass `max` to cap a key's quantity
 * on `add` (e.g. the still-unpaid amount of a position).
 */
export function useMengen<K extends string | number>(
  max?: (key: K) => number,
): MengenSteuerung<K> {
  const [mengen, setMengen] = useState<Record<K, number>>(
    () => ({}) as Record<K, number>,
  )

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
