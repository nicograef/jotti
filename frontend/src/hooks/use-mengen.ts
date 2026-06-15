import { useState } from 'react'

/**
 * Quantity-selector state keyed by id (variant id for ordering/direct sale,
 * position id for payment). `add` increments, `remove` decrements but never
 * below zero, `reset` clears the whole selection. Pass `max` to cap a key's
 * quantity (e.g. the still-unpaid amount of a position).
 */
export function useMengen<K extends string | number>(max?: (key: K) => number) {
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

  return { mengen, add, remove, reset }
}
