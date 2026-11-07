import { useCallback, useState } from 'react'

// Erfolgsmeldung einer Buchungsseite. `nachSchliessen` (der Refetch) läuft erst
// beim Schließen, damit sichtbare Statuswechsel dem Pop folgen; es muss eine
// stabile Referenz sein, sonst startet der Anzeige-Timer bei jedem Render neu.
export function useErfolgsPop(nachSchliessen: () => void): {
  zeigeErfolg: (nachricht: string) => void
  pop: { open: boolean; text: string; onDismiss: () => void }
} {
  const [erfolg, setErfolg] = useState({ open: false, text: '' })
  const zeigeErfolg = useCallback((nachricht: string) => {
    setErfolg({ open: true, text: nachricht })
  }, [])
  const onDismiss = useCallback(() => {
    setErfolg((prev) => ({ ...prev, open: false }))
    nachSchliessen()
  }, [nachSchliessen])
  return { zeigeErfolg, pop: { ...erfolg, onDismiss } }
}
