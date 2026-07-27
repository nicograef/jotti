import { renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { useVorgangId } from './use-vorgang-id'

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/

const nutzdaten = (menge: number, kommentar = '') => ({
  tischId: 1,
  positionen: [{ positionId: 'a', menge }],
  kommentar,
})

describe('useVorgangId', () => {
  it('behält den Schlüssel, solange die Nutzdaten gleich bleiben (auch bei neuer Objekt-Identität)', () => {
    const { result, rerender } = renderHook((props) => useVorgangId(props), {
      initialProps: nutzdaten(1),
    })

    const erster = result.current
    expect(erster).toMatch(UUID)

    rerender(nutzdaten(1))
    expect(result.current).toBe(erster)
  })

  it('erzeugt einen neuen Schlüssel, sobald sich die Nutzdaten ändern', () => {
    const { result, rerender } = renderHook((props) => useVorgangId(props), {
      initialProps: nutzdaten(1),
    })
    const erster = result.current

    // Geänderte Menge = geänderte Nutzdaten = neuer Vorgang.
    rerender(nutzdaten(2))
    const zweiter = result.current
    expect(zweiter).toMatch(UUID)
    expect(zweiter).not.toBe(erster)

    // Auch der Kommentar gehört zu den Nutzdaten.
    rerender(nutzdaten(2, 'ohne Zwiebeln'))
    expect(result.current).not.toBe(zweiter)
  })

  it('verwendet nach einer Änderung auch für frühere Nutzdaten keinen alten Schlüssel wieder', () => {
    const { result, rerender } = renderHook((props) => useVorgangId(props), {
      initialProps: nutzdaten(1),
    })
    const erster = result.current

    rerender(nutzdaten(2))
    const zweiter = result.current

    // Zurück zu den ursprünglichen Nutzdaten: ein neuer Vorgang, keine
    // Wiederverwendung des alten Schlüssels — der Server soll regulär prüfen.
    rerender(nutzdaten(1))
    expect(result.current).not.toBe(erster)
    expect(result.current).not.toBe(zweiter)
  })
})
